<?php
require_once '../cameo.php';

if($_SERVER['REQUEST_METHOD'] != 'POST') {
	http_response_code(500);
	exit();
}

if(!isset($_POST['target'])) {
	http_response_code(400);
	exit();
}

$target = trim(sanitize_target($_POST['target']));
email_result(null, "Error in Target: " . $target, "Error");
