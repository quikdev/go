<html>
	<head>
		<title>{{ .Name }} Browser Test</title>
		<meta charset="utf-8"/>
		<script src="wasm_exec.js"></script>
		<script>
			const go = new Go()

			WebAssembly.instantiateStreaming(fetch("{{ .Name }}.wasm"),
			go.importObject).then((result) => {
				go.run(result.instance);
				document.body.innerText = hello() // hello() comes from Go!
			})

			const url = new URL(location.href);
			url.pathname = "/livereload";

			const events = new EventSource(url.href);

			events.onmessage = function(event) {
				console.log(event)
				if (event.data === "reload") {
					console.warn("reloading!")
					location.reload();
				}
			}

			events.onerror = function(err) {
				console.error(err.message)
				events.close()
			}
		</script>
	</head>
	<body></body>
</html>