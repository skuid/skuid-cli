# Backwards Compatibility Docker Image

Hey y'all this is a docker image that's going to do a few things
1. clone `skuid-cli` version of the tool
1. clone `tides` version of the tool
1. call retrieve with `skuid-cli`
1. call retrieve with `tides`
1. go through and format all json documents (to eliminate newlines)
1. call a `diff` on both file directories

Here's how to run it
```bash
$> docker build . -t backwards-compatibility \
	 --build-arg HOST=${your-host-here} \
	 --build-arg USERNAME=${your-username-here} \
	 --build-arg PASSWORD=${your-password-here}
```

the build process is going to do everything but the diff. to see the diff, run
```bash
$> docker run -t backwards-compatibility
```