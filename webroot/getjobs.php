<?php
require_once '../cameo.php';

if($_SERVER['REQUEST_METHOD'] != 'POST') {
	http_response_code(500);
	exit();
}

$server = $_POST['SERVER'];
if(!in_array($server, $config['CAMEO']['servers'])) {
	http_response_code(500);
	exit();
}

mkdir(__DIR__ . "/../done/$server/", 0644, true);
header('Content-Type: text/plain');
foreach(new DirectoryIterator(__DIR__ . "/../jobs/$server/") as $file) {
	if($file->isFile() && $file->getExtension() == "json") {
		echo $file->getBasename() . ':' . base64_encode(file_get_contents($file)) . "\n";
		rename($jobfile, __DIR__ . "/../done/$server/" . $file->getBasename());
	}
}
