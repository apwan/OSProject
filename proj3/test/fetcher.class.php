<?php
class curl {
	public static $_UA = "Mozilla/4.0 (Windows; MSIE 7.0; Windows NT 5.1; SV1; .NET CLR 2.0.50727; MySimpleFetcher/PHP5 chenxq13#mails.tsinghua.edu.cn)";
	public static $_Cookie = "./cookie.txt";
	public static $_XFIP = "59.66.130.223";
	public static $_Curl_Handle = null;
	public static $_Curl_Handles = array();
	public static function buildCurlHandle($ch, $url, $ref=false)
	{	
		curl_setopt($ch, CURLOPT_URL, $url);
		curl_setopt($ch, CURLOPT_FAILONERROR, true);
		curl_setopt($ch, CURLOPT_FOLLOWLOCATION, true);
		curl_setopt($ch, CURLOPT_VERBOSE, 0);
		curl_setopt($ch, CURLOPT_HEADER, 0); //返回header部分?
		curl_setopt($ch, CURLOPT_RETURNTRANSFER, true); //返回字符串，而非直接输出
		
		curl_setopt($ch, CURLOPT_ENCODING, 'gzip,deflate');//解释gzip
		curl_setopt($ch, CURLOPT_TIMEOUT, 30); // 设置超时限制防止死循环
		
		if($ref){
			curl_setopt($ch, CURLOPT_REFERER, $ref);//带来的Referer
		}else{
			curl_setopt($ch, CURLOPT_AUTOREFERER, 1); // 自动设置Referer
		}
		
		
		$_Cookie=realpath(self::$_Cookie);
		if(!is_readable($_Cookie) || !is_writable($_Cookie)) die( 'No cookie saver.');
		
		curl_setopt($ch, CURLOPT_COOKIEJAR,  $_Cookie ); //存储cookies
		curl_setopt($ch, CURLOPT_COOKIEFILE, $_Cookie ); //存储cookies
		$head=array(
		 'X-FORWARDED-FOR:'.self::$_XFIP,
		 'User-Agent: '.self::$_UA,
		 'Accept: text/html,application/xhtml+xml',
		 'DNT: 1',
		 'Accept-Language: zh-cn',
		 'Accept-Encoding: gzip,deflate',
		 'Connection: Keep-Alive',
		 'Cache-Control: no-cache'
		);
		curl_setopt($ch,CURLOPT_HTTPHEADER,$head);
		return $ch;
	}
	public static function Get($url, $ref=false)//single request
	{
		//$ch = curl_init($url); //初始化
		
		if(! self::$_Curl_Handle) self::$_Curl_Handle=curl_init();
		$ch = self::$_Curl_Handle;
		$ch = self::buildCurlHandle($ch, $url, $ref);
		$data = curl_exec($ch);
		//curl_close($ch);	//connection reuse
		return $data;
	}
	
	public static function multiGet($urls, $ref=false)
	{
		$chs = self::$_Curl_Handles;
		$queue = curl_multi_init();
		foreach($urls as $i=>$url)
		{
			if(!isset($chs[$i]) || !$chs[$i])
				$chs[$i] = curl_init();

			$chs[$i] = self::buildCurlHandle($chs[$i], $url, $ref);
			curl_multi_add_handle($queue, $chs[$i]);
		}
		
		echo "cm_start\n";
		//start curl_multi
		$active = null;
		// execute the handles
		do {
			$mrc = curl_multi_exec($queue, $active);
			echo '.';
		} while ($mrc == CURLM_CALL_MULTI_PERFORM);

		while ($active > 0 && $mrc == CURLM_OK) {
			if (curl_multi_select($queue) != -1) {
				do {
					$mrc = curl_multi_exec($queue, $active);
					echo ',';
				} while ($mrc == CURLM_CALL_MULTI_PERFORM);
			}
			usleep(1000*1000*0.01);echo '>';
			$mrc = curl_multi_exec($queue, $active);
		}
		//done curl_multi

		$responses = array();
		foreach ($urls as $i=>$url) {
			$responses[$i] = curl_multi_getcontent($chs[$i]);
			curl_multi_remove_handle($queue, $chs[$i]);
			//curl_close($ch);
		}
		curl_multi_close($queue);
		
		self::$_Curl_Handles = $chs;
		return $responses;
	}
}
