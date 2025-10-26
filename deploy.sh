#!/usr/bin/env bash

# Использование:
# ./deploy.sh /path/to/project user@remote_host:/remote/path [SSH_KEY]

LOCAL_PATH=$1
REMOTE_PATH=$2
SSH_KEY=$3

if [ -z "$LOCAL_PATH" ] || [ -z "$REMOTE_PATH" ]; then
  echo "Использование: $0 <local_path> <user@remote_host:/remote_path> [ssh_key]"
  exit 1
fi

RSYNC_CMD="rsync -avz --delete --exclude={'.git','.idea'}"

if [ -n "$SSH_KEY" ]; then
  RSYNC_CMD="$RSYNC_CMD -e 'ssh -i $SSH_KEY'"
fi

echo "Синхронизация $LOCAL_PATH → $REMOTE_PATH ..."
eval "$RSYNC_CMD $LOCAL_PATH/ $REMOTE_PATH"

if [ $? -eq 0 ]; then
  echo "✅ Копирование завершено успешно."
else
  echo "❌ Ошибка при копировании."
  exit 1
fi
