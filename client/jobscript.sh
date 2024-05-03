#!/bin/bash -e
#SBATCH -J {{server}}-{{target}}-{{stoichiometry}}
#SBATCH -p gpu
#SBATCH --gres=gpu:1
#SBATCH -t 2-0
#SBATCH -c 4
#SBATCH -o {{server}}/{{target}}-{{stoichiometry}}.log

PASS_PAR=""
if [ "{{PASSWORD}}" != "" ]; then
    PASS_PAR="-F PASSWORD={{PASSWORD}}"
fi

err_report() {
    curl -X POST -F TARGET='{{target}}' ${PASS_PAR} '{{response}}/error'
}
trap 'err_report' ERR

WORKDIR="{{BASE}}/{{server}}/{{target}}-{{stoichiometry}}"
mkdir -p "${WORKDIR}"

hostname > "${WORKDIR}/job.host"
pwd >> "${WORKDIR}/job.host"

cat > "${WORKDIR}/job.csv" <<'End-of-Fasta'
id,sequence
{{target}},{{sequence}}
End-of-Fasta

colabfold_batch "${WORKDIR}/job.csv" "${WORKDIR}/result" {{COLABFOLD_PARAMS}}

RELAXED=relaxed
BEST=$(echo "${WORKDIR}/result/"*"_${RELAXED}_rank_001_"*".pdb")

# see https://predictioncenter.org/casp15/index.cgi?page=format
TARGET="{{target}}"
cat > "${WORKDIR}/job.pdb" <<End-of-Pdb-Header
PFRMAT TS
TARGET {{target}}
AUTHOR {{GROUP_ID}}
METHOD {{GROUP_SERVER}}
METHOD {{COLABFOLD_VERSION}}
METHOD ${PARAMS}
End-of-Pdb-Header

CNT=1
for i in {1..5}; do
    FILE="$(echo "${WORKDIR}/result/"*"_${RELAXED}_rank_00"${i}"_"*".pdb")"
    if [ ! -e "${FILE}" ]; then
      continue
    fi
    NAMES="$(echo "${WORKDIR}/result/"*"_template_domain_names.json")"
    if [ "${RELAXED}" = "relaxed" ]; then
        awk -v i=${CNT} 'BEGIN { printf("MODEL %8s\n", i); }' >> "${WORKDIR}/job.pdb"
    fi
    cat "${FILE}" | "{{BASE}}/add_parents.py" - "${NAMES}" "{{stoichiometry}}" \
        | awk -v i=${CNT} \
            '$1 == "END" { next; } \
            $1 == "MODEL" { printf("MODEL %8s\n", i); next; } \
            1; ' \
        >> "${WORKDIR}/job.pdb"
    if [ "${RELAXED}" = "relaxed" ]; then
        echo "ENDMDL" >> "${WORKDIR}/job.pdb"
    fi
    CNT=$((CNT+1))
done
echo "END" >> "${WORKDIR}/job.pdb"

#awk -v id=${GROUP_ID_HUMAN} -v orig=${GROUP_SERVER} -v server=${GROUP_SERVER_HUMAN} '$1 == "AUTHOR" { $2 = id; } $1 == "METHOD" && $2 == orig { $2 = server; } { print; }' "${WORKDIR}/job.pdb" > "${WORKDIR}/human.pdb"

if [ -s "${WORKDIR}/job.pdb" ]; then
    curl -X POST ${PASS_PAR} -F REPLY-E-MAIL='{{email}}' -F TARGET='{{target}}' -F SERVER='{{server}}' -F FILE=@"${WORKDIR}/job.pdb" '{{response}}/success'
    #curl -X POST -F REPLY-E-MAIL='models@predictioncenter.org' -F TARGET='{{target}}' -F SERVER=${GROUP_SERVER_HUMAN} -F FILE=@"${WORKDIR}/human.pdb" '{{response}}/success'
else
    curl -X POST ${PASS_PAR} -F TARGET='{{target}}' '{{response}}/error'
fi

cp -f -- "${WORKDIR}/job.pdb" "{{UPLOAD}}/{{target}}-{{stoichiometry}}.pdb"
chmod a+r "{{UPLOAD}}/{{target}}-{{stoichiometry}}.pdb"
tar -C "{{BASE}}/{{server}}" -czvf "{{UPLOAD}}/{{target}}-{{stoichiometry}}.tar.gz" "{{target}}-{{stoichiometry}}/result"
chmod a+r "{{UPLOAD}}/{{target}}-{{stoichiometry}}.tar.gz"
cp -f -- "${WORKDIR}/result/{{target}}_coverage.png" "{{UPLOAD}}/{{target}}-{{stoichiometry}}.png" 
chmod a+r "{{UPLOAD}}/{{target}}-{{stoichiometry}}.png"
cp -f -- "${WORKDIR}/result/{{target}}_coverage.png" "{{UPLOAD}}/{{target}}-{{stoichiometry}}_coverage.png" 
chmod a+r "{{UPLOAD}}/{{target}}-{{stoichiometry}}_coverage.png"
cp -f -- "${WORKDIR}/result/{{target}}_plddt.png" "{{UPLOAD}}/{{target}}-{{stoichiometry}}_plddt.png" 
chmod a+r "{{UPLOAD}}/{{target}}-{{stoichiometry}}_plddt.png"
cp -f -- "${WORKDIR}/result/{{target}}_pae.png" "{{UPLOAD}}/{{target}}-{{stoichiometry}}_pae.png" 
chmod a+r "{{UPLOAD}}/{{target}}-{{stoichiometry}}_pae.png"
(cd "{{UPLOAD}}" && bash "./make_index.sh" > "./index.html")
