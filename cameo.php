<?php
error_reporting(E_ALL);
ini_set("display_errors", 1);

require __DIR__ . '/vendor/autoload.php';
$config = parse_ini_file(__DIR__ . 'settings.ini', true);

if($_POST['KEY'] != $config['CAMEO']['key']) {
	http_response_code(500);
	exit();
}

function email_result($email, $subject, $body) {
	$mail = new PHPMailer\PHPMailer\PHPMailer(true);

	$mail->isSMTP();
	$mail->SMTPAuth = true;
	$mail->Host = $config['SMTP']['host'];
	$mail->Port = $config['SMTP']['port'];
	$mail->Username = $config['SMTP']['username'];
	$mail->Password = $config['SMTP']['password'];
	$mail->SMTPSecure = $config['SMTP']['securemode'];
	$mail->AllowEmpty = true;

	$mail->setFrom($config['SMTP']['from']);
	if ($email != null) {
		$mail->addAddress($email);
	}
	if (isset($config['BCC'])) {
		foreach($config['BCC']['bcc'] as $value) {
			$mail->addBCC($value);
		}
	}

	$mail->Subject = $subject;
	$mail->Body = $body;

	$mail->send();

	return true;
}

function sanitize_target($target) {
	return preg_replace('/[^A-Za-z0-9\-\_]/', '', $target);
}

function sanitize_sequence($sequence) {
	return preg_replace('/[^A-Z]/', '', strtoupper($sequence));
}
