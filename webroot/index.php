<?php
require_once '../cameo.php';

if($_SERVER['REQUEST_METHOD'] != 'POST') {
	http_response_code(500);
	exit();
}

if(!isset($_POST['TARGET']) || !isset($_POST['SEQUENCE']) || !isset($_POST['REPLY-E-MAIL']) || !isset($_POST['SERVER'])) {
	http_response_code(400);
	exit();
}

$email = $_POST['REPLY-E-MAIL'];
if(!filter_var($email, FILTER_VALIDATE_EMAIL)) {
	http_response_code(400);
	exit();
}

$target = sanitize_target($_POST['TARGET']);
$sequence = sanitize_sequence($_POST['SEQUENCE']);

$server = $_POST['SERVER'];
if(!in_array($server, $config['CAMEO']['servers'])) {
	http_response_code(500);
	exit();
}

mkdir(__DIR__ . "/../jobs/$server/", 0644, true);
file_put_contents(__DIR__ . "/../jobs/${server}/${target}.json", json_encode(array(
	'server' => $server,
	'target' => $target,
	'sequence' => $sequence,
	'email' => $email,
	'host' => $config['CAMEO']['host']
))));

/*
$confirmationAddress = null;
if(isset($config[$server]) && isset($config[$server]['confirmation'])) {
	$confirmationAddress = $config[$server]['confirmation'];
}
if($confirmationAddress != null) {
	email_result($confirmationAddress, "$target - query received by $server", "");
}
*/