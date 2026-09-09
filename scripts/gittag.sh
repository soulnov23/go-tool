#!/bin/bash

# 不设置set -x，避免执行过程干扰日志输出，需要排查时执行bash -x scripts/gittag.sh
set -e

# v${MAJOR}.${MINOR}.${PATCH}
TAG_PREFIX="v"
INITIAL_VERSION="1.1.1"
MINOR_CARRY_THRESHOLD=99
PATCH_CARRY_THRESHOLD=99

function log() {
    printf '\033[1;32m[gittag]\033[m %s\n' "$*" >&2
}

# 获取当前HEAD commit上的所有tag
function head_tags() {
    # git失败时管道退出码取自grep，视为无标签
    git tag --points-at HEAD | grep "^${TAG_PREFIX}" || true
}

# 解析MAJOR.MINOR.PATCH，忽略-及之后的后缀（如1.2.3-rc1），忽略第三段之后的多余段（如1.2.3.4）
# 解析结果写入全局MAJOR、MINOR、PATCH
function parse_version() {
    if [[ ! $1 =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)([.-].*)?$ ]]; then
        return 1
    fi
    # 10#强制按十进制解析，避免08、09之类前导零被当作八进制
    MAJOR=$((10#${BASH_REMATCH[1]}))
    MINOR=$((10#${BASH_REMATCH[2]}))
    PATCH=$((10#${BASH_REMATCH[3]}))
}

# 获取当前版本，无有效标签时写入初始版本并置IS_INITIAL为1
function current_version() {
    local tag version
    # --sort=-version:refname按版本号倒序，取第一个能解析的标签
    while read -r tag; do
        if [ -z "${tag}" ]; then
            continue
        fi
        version=${tag#"${TAG_PREFIX}"}
        if [ "${version}" = "${tag}" ]; then
            log "📢 跳过无效标签[${tag}]: 缺少前缀[${TAG_PREFIX}]"
            continue
        fi
        if ! parse_version "${version}"; then
            log "📢 跳过无效标签[${tag}]: 无效的版本格式[${version}]"
            continue
        fi
        log "✅ 从Git标签获取当前版本[${version}]"
        CURRENT_VERSION=${version}
        IS_INITIAL=0
        return
    done < <(git tag --sort=-version:refname)

    log "📢 未找到任何有效Git标签，使用初始版本[${INITIAL_VERSION}]"
    CURRENT_VERSION=${INITIAL_VERSION}
    IS_INITIAL=1
}

# 递增版本号，结果写入全局NEW_VERSION
function increment_version() {
    if ! parse_version "$1"; then
        log "❌ 无效的版本格式[$1]"
        exit 1
    fi

    # 智能进位逻辑
    PATCH=$((PATCH + 1))
    if [ ${PATCH} -gt ${PATCH_CARRY_THRESHOLD} ]; then
        PATCH=0
        MINOR=$((MINOR + 1))
        if [ ${MINOR} -gt ${MINOR_CARRY_THRESHOLD} ]; then
            MINOR=0
            MAJOR=$((MAJOR + 1))
        fi
    fi
    NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
    log "✅ 版本递增成功[$1->${NEW_VERSION}]"
}

# 创建并推送Git标签，伪原子操作：push失败则回滚本地tag
function create_and_push_git_tag() {
    local tag=${TAG_PREFIX}$1

    if ! git tag "${tag}"; then
        log "❌ 创建Git标签[${tag}]失败"
        exit 1
    fi
    log "✅ 已创建Git标签[${tag}]"

    if ! git push origin "${tag}"; then
        log "❌ 推送Git标签[${tag}]失败"
        if git tag -d "${tag}"; then
            log "📢 已回滚本地标签[${tag}]"
        else
            log "📢 回滚本地标签[${tag}]失败"
            log "📢 请手动执行[git tag -d ${tag}]"
        fi
        exit 1
    fi
    log "✅ 已推送Git标签[${tag}]"
}

main() {
    # 1. 检查当前HEAD是否已经打过tag，如果是则跳过
    local tags
    tags=$(head_tags)
    if [ -n "${tags}" ]; then
        log "📢 当前最新提交已有标签[${tags//$'\n'/ }]，跳过打标签"
        exit 1
    fi

    # 2. 获取当前版本
    current_version

    # 3. 递增版本号（初始版本不递增）
    NEW_VERSION=${CURRENT_VERSION}
    if [ ${IS_INITIAL} -eq 0 ]; then
        increment_version "${CURRENT_VERSION}"
    fi

    # 4. 创建并推送Git标签
    create_and_push_git_tag "${NEW_VERSION}"
}

main "$@"
