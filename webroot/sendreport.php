<?php
require_once '../cameo.php';

if($_SERVER['REQUEST_METHOD'] != 'POST') {
	http_response_code(500);
	exit();
}

if(!isset($_POST['target']) || !isset($_POST['email']) || !isset($_POST['server']) || !isset($_FILES['file'])) {
	http_response_code(400);
	exit();
}

$email = $_POST['email'];
if(!filter_var($email, FILTER_VALIDATE_EMAIL)) {
	http_response_code(400);
	exit();
}

$server = $_POST['server'];
if(!in_array($server, $config['CAMEO']['servers'])) {
	http_response_code(500);
	exit();
}

$target = trim(sanitize_target($_POST['target']));

email_result($email, $target . ' - ' . $server, file_get_contents($_FILES['file']['tmp_name']));
