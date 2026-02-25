#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# add_service.sh
# Добавляет новый сервис: proto/<service>/<version>/<service>.proto
# Использование:
#   ./script/add_service.sh users
#   ./script/add_service.sh users v2
#   make add-service users
#   make add-service users v2
# =============================================================================

if [ $# -lt 1 ] || [ $# -gt 2 ]; then
    echo "Ошибка: ожидается 1 или 2 аргумента"
    echo "Примеры:"
    echo "  $0 users"
    echo "  $0 users v2"
    echo "  make add-service users"
    echo "  make add-service users v2"
    exit 1
fi

SERVICE_NAME="$1"
VERSION="${2:-v1}"                        # по умолчанию v1

SERVICE_LOWER=$(echo "$SERVICE_NAME" | tr '[:upper:]' '[:lower:]')
VERSION_LOWER=$(echo "$VERSION" | tr '[:upper:]' '[:lower:]')

# Путь к proto-файлу
PROTO_DIR="proto/${SERVICE_LOWER}/${VERSION_LOWER}"
PROTO_FILE="${PROTO_DIR}/${SERVICE_LOWER}.proto"

mkdir -p "${PROTO_DIR}"

if [ -f "${PROTO_FILE}" ]; then
    echo "Файл уже существует → ${PROTO_FILE}"
    echo "Пропускаем создание."
    exit 0
fi

# Проверяем наличие шаблона
TEMPLATE="templates/service.proto.tpl"
if [ ! -f "${TEMPLATE}" ]; then
    echo "Ошибка: шаблон не найден → ${TEMPLATE}"
    exit 1
fi

# Подставляем переменные
export SERVICE_NAME_PASCAL="${SERVICE_NAME}"
export SERVICE_NAME_LOWER="${SERVICE_LOWER}"
export PROTO_VERSION="${VERSION_LOWER}"
export MODULE="${MODULE:-github.com/oulabla/go-base}"   # fallback, если не задан в .project.mk или окружении

envsubst < "${TEMPLATE}" > "${PROTO_FILE}"

echo "Создан новый сервис:"
echo "  → ${PROTO_FILE}"
echo "     (версия: ${PROTO_VERSION})"
echo ""
echo "Следующие шаги:"
echo "  make generate"
echo "  # или"
echo "  make proto-generate && make scaffold"