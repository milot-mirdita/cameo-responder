#!/bin/sh -e
source ./config.sh

WORKDIR="${BASE}/api"
mkdir -p "${WORKDIR}"

INPUT="$1"
if [ -e "${INPUT}" ]; then
    API="$1"
else
    TS=$(date +%s)
    API="${WORKDIR}/api-${TS}"
    curl -S -s -X POST https://casp.colabfold.com/jobs > "${API}"
    if [ ! -s "${API}" ]; then
        rm -f -- "${API}"
        exit 0
    fi
fi

while read -r line; do
    eval $(echo "$line" | jq -r "to_entries|map(\"export \(.key)=\(.value|tostring)\"),map(\"export \(.key)_len=\(.value|length)\")|.[]") 
    mkdir -p "${BASE}/${server}/${target}-${stoichiometry}"
    ./mo jobscript.sh > "${BASE}/${server}/${target}-${stoichiometry}/job.sh"
    sbatch -Q "${BASE}/${server}/${target}-${stoichiometry}/job.sh"
done < "${API}"
