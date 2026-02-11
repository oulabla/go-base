#!/bin/bash
set -euo pipefail

mkdir -p internal/endpoints

camel_to_snake() {
    echo "$1" | sed -r 's/([A-Z])/_\L\1/g' | sed 's/^_//'
}

if [ -z "${MODULE:-}" ] || [ -z "${PROJECT_NAME:-}" ]; then
    echo "Error: MODULE and/or PROJECT_NAME not set. Run 'make init' first."
    exit 1
fi

# Объявляем массив СНАРУЖИ циклов
declare -a services=()

# Собираем все proto-файлы один раз
proto_files=$(find proto -name '*.proto' | sort)

for proto_file in $proto_files; do
    # Получаем все сервисы из файла (без подпроцесса while)
    services_in_file=$(grep -E '^service ' "$proto_file" | awk '{print $2}')

    for svc in $services_in_file; do
        svc_snake=$(camel_to_snake "$svc")
        
        # Убираем "service"/"Service" из конца
        svc_pkg="${svc_snake}"
        svc_pkg="${svc_pkg%_service}"
        svc_pkg="${svc_pkg%service}"

        endpoints_dir="internal/endpoints/${svc_pkg}"
        mkdir -p "$endpoints_dir"

        # Запоминаем сервис (формат: pkg|original_name)
        services+=("${svc_pkg}|${svc}")

        echo "Processed service: ${svc} → package ${svc_pkg}"

        # 1. Controller (только если нет файла)
        controller_file="${endpoints_dir}/controller.go"
        if [ ! -f "$controller_file" ]; then
            export SERVICE_PKG_NAME="$svc_pkg"
            export SERVICE_NAME="$svc"
            envsubst < templates/controller.go.tpl > "$controller_file"
            echo "  Created controller   : $controller_file"
        fi

        # 2. Методы (по одному файлу на rpc)
        # Извлекаем rpc-методы без вложенного подпроцесса
        methods=$(awk "/service ${svc}/,/}/" "$proto_file" | grep -E '^[[:space:]]*rpc ' | \
                  sed -E 's/^[[:space:]]*rpc ([A-Za-z0-9_]+).*/\1/')

        for method in $methods; do
            [ -z "$method" ] && continue

            method_snake=$(camel_to_snake "$method")
            method_file="${endpoints_dir}/${method_snake}.go"

            if [ ! -f "$method_file" ]; then
                export SERVICE_PKG_NAME="$svc_pkg"
                export METHOD_NAME="$method"
                envsubst < templates/endpoint_method.go.tpl > "$method_file"
                echo "  Created endpoint     : $method_file"
            fi
        done
    done
done

echo "Endpoints generation finished."