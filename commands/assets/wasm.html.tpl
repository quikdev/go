<html>
	<head>
		<meta charset="utf-8"/>
		<script src="wasm_exec.js"></script>
		<script>
			const go = new Go();
			WebAssembly.instantiateStreaming(fetch("example.wasm"),
			go.importObject).then((result) => {
				go.run(result.instance);
				document.body.innerText = hello() // hello() comes from Go!
			});
		</script>
	</head>
	<body></body>
</html>