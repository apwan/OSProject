<?php
date_default_timezone_set('Asia/Shanghai');
require 'fetcher.class.php';
//check
$conf=json_decode(file_get_contents('../conf/settings.conf'),1);
var_dump($conf);
$kv_url='http://'.$conf['primary'].':'.$conf['port'].'/kv/';

curl::Post($kv_url.'insert',array('key'=>'!','value'=>'#'));
$raw1=curl::Get($kv_url.'get',array('key'=>'!'));
$ret1=json_decode($raw1,1);
if($ret1['success']!=true ||$ret1['value']!='#')die('Is the server up and running?\n'.raw1);

$N=1000;
$urls=array();
$dataset=array();
for($i=0;$i<$N;$i++)
{
	$urls[$i]=$kv_url.'update';
	$dataset[$i]=array('key'=>'$a'.$i,'value'=>'vca'.$i);
}
$time_start = microtime(true);
$ret1=curl::Multi($urls,$dataset,true);
$ret2=curl::Multi($urls,$dataset,false);
$time_end = microtime(true);
$time = $time_end - $time_start;
echo "Performed $N R/W in time:".$time;