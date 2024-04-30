#!/bin/sh -xe
. ./config.sh
export BASE
export UPLOAD
export COLABFOLD_VERSION
export GROUP_ID
export GROUP_SERVER
export PASSWORD

WORKDIR="${BASE}/api"
mkdir -p "${WORKDIR}"

PASS_PAR=""
if [ "${PASSWORD}" != "" ]; then
    PASS_PAR="-F PASSWORD=${PASSWORD}"
fi

INPUT="$1"
if [ -e "${INPUT}" ]; then
    API="$1"
else
    TS=$(date +%s)
    API="${WORKDIR}/api-${TS}"
    curl -S -s -X POST ${PASS_PAR} "${ENDPOINT}" > "${API}"
    if [ ! -s "${API}" ]; then
        rm -f -- "${API}"
        exit 0
    fi
fi

while read -r line; do
    eval $(echo "$line" | jq -r "to_entries|map(\"export \(.key)=\(.value|tostring)\"),map(\"export \(.key)_len=\(.value|length)\")|.[]") 
    if [ -e "./config.sh.d/${server}" ]; then
        . "./config.sh.d/${server}"
        export UPLOAD
        export COLABFOLD_VERSION
        export GROUP_ID
        export GROUP_SERVER
        export PASSWORD
    fi
    mkdir -p "${BASE}/${server}/${target}-${stoichiometry}"
    ./mo jobscript.sh > "${BASE}/${server}/${target}-${stoichiometry}/job.sh"
    sbatch -D "$BASE" -Q "${BASE}/${server}/${target}-${stoichiometry}/job.sh"
done < "${API}"
