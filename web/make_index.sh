#!/bin/bash -e

DIR="."
# Gather all .pdb files in the directory
files=()
for file in "$DIR"/*.pdb; do
    files+=("$(basename "$file")")
done

# Sort the files by version number
IFS=$'\n' files=($(sort -Vr <<<"${files[*]}"))
unset IFS

# Get the latest file
latest="${files[0]}"

# Start generating HTML
cat <<-EOF
<!doctype html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1,maximum-scale=1.0,user-scalable=no">
<title>ColabFold CASP15 Predictions</title>
<link rel="apple-touch-icon" sizes="180x180" href="apple-touch-icon.png">
<link rel="icon" type="image/png" sizes="32x32" href="favicon-32x32.png">
<link rel="icon" type="image/png" sizes="16x16" href="favicon-16x16.png">
<link rel="shortcut icon" href="favicon.ico">
<style>
* {
    box-sizing: border-box;
}
body {
    font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
}

@media screen and (max-width:767px) {
#ngl {
    position: absolute;
    right: 0;
    top: 0;
    left: 0;
    height: 66vh;
}
#name {
    position: absolute;
    right: 15px;
    top: 15px;
}
#models {
    position: absolute;
    right: 15px;
    top: calc(66vh - 45px);
    touch-action: none;
}
#list {
    margin-top: calc(66vh + 2em);
}
}
@media screen and (min-width:768px) {
#ngl {
    position: fixed;
    right: 0;
    top: 0;
    height: 100vh;
    width: 50vw;
}
#name {
    position: fixed;
    right: 15px;
    top: 20px;
}
#models {
    position: fixed;
    left: calc(50vw + 15px);
    top: 15px;
}
#list {
    max-width: 50vw;
}
}
#ngl {
    background-color:black;
}
#name {
    font-size: 24px;
    color: #eee;
}
button {
    all: initial;
    box-sizing: border-box;
    background-color: #eee;
    background-repeat: no-repeat;
    background-position: 4px 45%;
    padding: 10px 15px;
    vertical-align: middle;
    border: 1px solid #ddd;
    border-radius: 3px;
    cursor: pointer;
    font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
    font-size: 14px;
    font-weight: bold;
    line-height: 15px;
}
a {
    text-decoration: none;
}
button:hover {
    box-shadow: 0 0 1px black;
}
button.active {
    background-color: #EF9250;
}
.button-group {
    display:inline-flex;
    flex-direction:row;
    vertical-align: middle;
}
.button-group button {
    border-radius:0;
}
.button-group button:first-child {
    border-top-left-radius: 3px;
    border-bottom-left-radius: 3px;
}
.button-group button:last-child {
    border-top-right-radius: 3px;
    border-bottom-right-radius: 3px;
}
#list {
    padding-left: 14px;
}
section {
    padding: 14px 0;
}
section img {
    padding: 5px;
    max-width: calc(100% - 10px);
}
h1 {
    padding-bottom: 0;
    margin-bottom: 0;
    background-image:url(apple-touch-icon.png);
    background-repeat:no-repeat;
    background-size: 1.5em;
    line-height:1.5em;
    background-position: center left;
    padding-left: 1.75em;
}
.foldseek {
    background-image: url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAFBUExURUdwTE9LSdLJwgAAAAAAAAAAAAgICAAAAAAAAAAAABkIABISDB0YGDgUAJ9DAg0AAAkAAEwhAOiGAgoKCh4KAKZHAg0AABMGBhEFBQgICBQUFDNNGQAAAAAAAA8AAEIIIXVtarYiSSsgII+ADzATABMDE45LJgAAABwcHFcRI0kMQrofRyY+E2p7CS4kBA8AAEWFI30XDkxlFScQCVkhZB4DAwsWBScFAjQHOkAJBF4YKg8AD1MmTw8HBxlcTl6PMxMkGUYJBXsnHMJ4Y2oPCa8lFX8RCigGAyEEAhEcHz2Ca05JR4OkeEE8Q1EcGZYTDUqNQz8lFEFCRGBwQEAJBhUAAAwAAGogFptZQiFgaIRpNGJFU5QgEQAAAP65AfSfAtmsB8C4H88rU9V/b8ipadAeE5tAIp2BSKGEX55SMN6LUVhh74UAAABedFJOUwCD+wcDEh4BCAUeKjQ/qxMaU+8YMrI3KCwfJooJDjCZo/qU8TVCug81j7z5u/toMevHzk2/Qy1bg2q6Mdsi6PlaisLxo/fAUnlY4Ij8y4Te+mW77m4kPL3S1fj11wKYhryCAAAAlklEQVQY02NgIBPoszNwcAJpNlYI39LN3VBNhoGBh4mRHcS3dc5KdjVmkBMT4AbzWZxS0jNSHU3VJWEmBEZlpsX4u5grCMNEwjxCwiNCGTSkQBxZTSvfIG+/WE8HM4isuKJ2UoCPV6SdkZ4KC1hEQl7LJjrY3kA3MV6UH6KIg9PawkRHOSFOiI8L7lZVJWkRQV5mYrwFAMN2FkvZZAq4AAAAAElFTkSuQmCC);
    background-repeat: no-repeat;
    background-position: 4px 45%;
    padding-left: 24px;
}
</style>
</head>
<body>
<div id="list">
<h1><a href="https://github.com/sokrypton/ColabFold">ColabFold</a> CASP15 Predictions</h1>
EOF

# Loop through files and generate HTML sections
for file in "${files[@]}"; do
    base="${file%.pdb}"

    cat <<-EOF
    <section>
        <h2>${base}</h2>
        <a download href="${base}.pdb"><button>PDB File</button></a>
        <a download href="${base}.tar.gz"><button>MSAs, etc.</button></a>
        <a href="#${base}.pdb"><button class="viz ${base}">Visualize</button></a>
        <img src="${base}.png" alt="MSA coverage">
    </section>
EOF

done

cat <<-'EOF'
<br>
<small><a href="make_index.sh" style="color:#aaa;">Show code</a></small>
</div>
<div id="ngl"></div>
<div id="name"></div>
<div id="models">
<button class="foldseek">Foldseek</button>
<div class="button-group">
<button class="model model-1" onclick="model(1)">1</button>
<button class="model model-2" onclick="model(2)">2</button>
<button class="model model-3" onclick="model(3)">3</button>
<button class="model model-4" onclick="model(4)">4</button>
<button class="model model-5" onclick="model(5)">5</button>
</div>
<button id="color" onclick="nextColor()" style="width:90px;text-align:center;">pLDDT</button>
</div>
<script src="./ngl.js"></script>
<script>
var $ngl = document.getElementById("ngl");
var $name = document.getElementById("name");
var $color = document.getElementById("color");

var stage = new NGL.Stage($ngl, {ambientIntensity:0.5, fogNear: 60});
var repr = null;
var comp = null;
var mdl = 0;

var loadfile = function(file) {
    stage.removeAllComponents();
    $name.innerText = file;
    var base = file.split(".")[0];
    document.querySelectorAll('.viz').forEach(function (element) {
        element.classList.remove('active');
    });
    document.querySelector('.viz.' + base).classList.add('active');
    fetch(file)
        .then(response => response.blob())
        .then(blob => {
           return stage.loadFile(blob, {ext: 'pdb'});
        })
        .then(pdb => {
            repr = pdb.addRepresentation("cartoon", { color: 'bfactor' });
            comp = pdb;
            model(1);
            return pdb;
        })
        .then(r => {
            var s = r.structure
            var base = s.getView(new NGL.Selection("/0 and .CA"))
            for (var i = 1; i <= r.structure.modelStore.count; i++) {
                var view = s.getView(new NGL.Selection(`/${i-1} and .CA`))
                var superposition = new NGL.Superposition(view, base)
                superposition.transform(view);
            }
        })

}

function nextColor() {
    var color = "bfactor";
    if ($color.innerText == "pLDDT") {
        $color.innerText = "Chain";
        color = "chainindex";
    } else if ($color.innerText == "Chain") {
        $color.innerText = "Rainbow";
        color = "atomindex";
    } else if ($color.innerText == "Rainbow") {
        $color.innerText = "pLDDT";
        color = "bfactor";
    }
    repr.setParameters({ colorScheme: color })
}

window.addEventListener("hashchange", event => {
    var target = location.hash.substr(1);
    loadfile(target);
}, false);

window.addEventListener('resize', function(event) {
    stage.handleResize();
}, true);

function ready(cb) {
  if (document.readyState === "loading") {
    document.addEventListener('DOMContentLoaded', cb);
  } else {
    cb();
  }
}

function model(num) {
    mdl = num - 1;
    repr.setSelection("/" + mdl)
    stage.autoView()
    document.querySelectorAll('.model').forEach(function (element) {
        element.classList.remove('active');
    });
    document.querySelector('.model-' + num).classList.add('active');
}
EOF

cat <<EOF
ready(() => {
    if (location.hash.length > 1) {
        var target = location.hash.substr(1);
        loadfile(target);
    } else {
        location.hash = "#$(printf "%s" "$latest")";
    }
});
EOF

cat <<-'EOF'
</script>
<script>
function convertToQueryUrl(obj) {
  var params = new URLSearchParams(obj);
  var entries = Object.entries(obj);

  for (var entry in entries) {
    var key = entries[entry][0];
    var value = entries[entry][1];

    if (Array.isArray(value)) {
      params.delete(key);
      value.forEach(function (v) {
        return params.append(key + '[]', v);
      });
    }
  }

  return params.toString();
}

function request(method, url, body) {
  return new Promise(function (resolve, reject) {
    var xhr = new XMLHttpRequest();

    xhr.onload = function () {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve(xhr.responseText);
      }

      reject([xhr.status, xhr.statusText]);
    };

    xhr.open(method, url);
    xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');

    if (typeof body != "undefined") {
      xhr.send(convertToQueryUrl(body));
    } else {
      xhr.send();
    }
  });
}

function ready(cb) {
  if (document.readyState === "loading") {
    document.addEventListener('DOMContentLoaded', cb);
  } else {
    cb();
  }
}

var template = "<!doctype html>\n<head>\n<meta name=\"viewport\" content=\"width=device-width,initial-scale=1,maximum-scale=1.0,user-scalable=no\">\n<title>Loading Foldseek</title>\n<style>\n  body {\n    background-color: #121212;\n    color: #fff;\n    font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;\n    height: 100vh;\n    display: flex;\n    flex-direction: column;\n    flex-wrap: wrap;\n    justify-content: center;\n    align-items: center;\n  }\n  .loader {\n    display: block;\n    width: 80px;\n    height: 80px;\n  }\n  .loader:after {\n    content: \" \";\n    display: block;\n    width: 64px;\n    height: 64px;\n    margin: 8px;\n    border-radius: 50%;\n    border: 6px solid #fff;\n    border-color: #fff transparent #fff transparent;\n    animation: loader 1.2s linear infinite;\n  }\n  @keyframes loader {\n    0% {\n      transform: rotate(0deg);\n    }\n    100% {\n      transform: rotate(360deg);\n    }\n  }\n</style>\n</head>\n<body>\n<div>Foldseek is loading...</div><div class=\"loader\"></div>\n</body>";
ready(function () {
  document.addEventListener('click', function (e) {
    if (e.target && e.target.classList.contains('foldseek')) {
      var w = window.open('', '_blank');
      w.document.body.innerHTML = template;
      var pdbUrl = location.hash.substr(1);

      if (typeof pdbUrl == "undefined") {
        console.warn("no url");
        return;
      }

      request('GET', pdbUrl).then(function (pdb) {
        var modelpdb = "";
        var lines = pdb.split('\n');
        var inModel = false;
        for (var i = 0; i < lines.length; i++) {
            if (lines[i].startsWith("MODEL") && lines[i].split(/\s+/)[1] == "1") {
                inModel = true;
            }
            if (lines[i].startsWith("ENDMDL")) {
                inModel = false;
            } 
            if (inModel && lines[i].startsWith("ATOM")) {
                modelpdb += lines[i];
                modelpdb += "\n";
            }
        }
        return request('POST', 'https://search.foldseek.com/api/ticket', {
          q: modelpdb,
          database: ["afdb-proteome", "afdb-swissprot", "gmgcl_id", "pdb100"],
          mode: "3diaa"
        });
      }).then(function (data) {
        w.location = 'https://search.foldseek.com/queue/' + JSON.parse(data).id;
      }).catch(function (error) {
        w.close();
      });
    }
  });
});
</script>
</body>
EOF
