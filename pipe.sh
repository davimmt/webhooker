#!/bin/bash

PIPELINE_ARGS=$@

mkdir -p "$GIT_ROOT_FOLDER"
cd "$GIT_ROOT_FOLDER"

git config --global core.compression 0

runPipe() {
  for git_repo in $(echo "$GIT_REPOS_TO_CLONE" | tr ',' '\n'); do
    repo_name="$(basename $git_repo .git)";
    rm -rf "$GIT_ROOT_FOLDER/$repo_name";
    git clone --depth 1 "$git_repo" "$GIT_ROOT_FOLDER/$repo_name";
  done

  "$GIT_ROOT_FOLDER/$PIPELINE_SCRIPT_PATH" $PIPELINE_ARGS
  rm -f $1
}

makeLock() {
  shopt -s nullglob;
  local lock_files=(lock.*)
  local lock="lock.0"
  if [[ "${#lock_files[@]}" -gt 0 ]]; then
    local last_lock=${lock_files[-1]};
    local next_lock=$(( $(echo $last_lock | awk -F. '{print $2}') + 1));
    local lock="lock.$next_lock";
    touch "$lock" && chmod a=r "$lock";

    while [[ -f $last_lock ]]; do sleep 1; done;
  else
    touch "$lock" && chmod a=r "$lock";
  fi
  echo $lock
}

lock=$(makeLock)
runPipe $lock
exit 0