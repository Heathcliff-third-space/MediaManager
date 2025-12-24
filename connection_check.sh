#!/bin/bash

echo "环境变量检查:"
echo "=============="

echo "TELEGRAM_BOT_TOKEN: ${TELEGRAM_BOT_TOKEN:-未设置}"
echo "AUDIOBOOKSHELF_TOKEN: ${AUDIOBOOKSHELF_TOKEN:-未设置}"
echo "AUDIOBOOKSHELF_URL: ${AUDIOBOOKSHELF_URL:-未设置}"
echo "PROXY_ADDRESS: ${PROXY_ADDRESS:-未设置}"

echo ""
echo "如果以上变量都正确显示，说明 .env 文件配置成功。"