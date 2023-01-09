#!/bin/bash

# Variables:
export LACKEY_LIBRARY_DIR="${LACKEY_LIBRARY_DIR-/mnt/hifi}"
export LACKEY_OUTPUT_DIR="${LACKEY_OUTPUT_DIR-/mnt/lofi}"
export LACKEY_SYNC_ARGS="${LACKEY_SYNC_ARGS-"-s -m -d -r 192k --opus --threshold 192"}"
export LACKEY_LOG_FILE="${LACKEY_LOG_FILE-/tmp/lackey.log}"

main() {
    lackey version | head -2

    echo
    echo "Current settings:"
    echo
    echo "  LACKEY_LIBRARY_DIR=${LACKEY_LIBRARY_DIR}"
    echo "  LACKEY_OUTPUT_DIR=${LACKEY_OUTPUT_DIR}"
    echo "  LACKEY_SYNC_ARGS=${LACKEY_SYNC_ARGS}"
    echo "  LACKEY_LOG_FILE=${LACKEY_LOG_FILE}"
    echo

    pid=$(pgrep lackey)
    if [ "${pid}" != "" ]; then
        echo "---"
        echo "Found lackey process running: ${pid}"
        echo "Replaying log file and following."
        echo "..."
        watch_log $pid
    else
        if [ -f ${LACKEY_LOG_FILE} ]; then
            echo "---"
            echo "Found lackey log from: " $(date -r ${LACKEY_LOG_FILE} +"%F %T")
            read -p "Replay log file? [Yn] " -n1 answer
            echo
            echo "..."
            if [ "${answer}" != "n" ]; then
                cat ${LACKEY_LOG_FILE}
            fi
            unset answer
            echo "---"
            echo
        fi

        read -p "Start new scan? [yN] " -n1 answer
        echo
        if [ "${answer}" != "y" ]; then
            echo "Abort."
            return
        fi

        countdown 5
        nohup lackey --color=always -L "${LACKEY_LIBRARY_DIR}" sync ${LACKEY_SYNC_ARGS} "${LACKEY_OUTPUT_DIR}" > ${LACKEY_LOG_FILE} 2>&1 &
        watch_log $!
    fi
}

countdown() {
    local n=${1-10}
    echo
    echo "Commencing in T–${n} seconds..."
    n=$((n - 1))
    for i in $(seq $n -1 1); do
        sleep 1
        echo "  $i ..."
    done
    echo "Commence."
    echo
}

watch_log() {
    local pid=$1

    tail --pid=$pid -n +1 -f ${LACKEY_LOG_FILE}
}

main
exit 0
