# LDT

LDT retrieves the latest tag, out of list of tags, retrieved from a custom docker repository, while matching the tag by prefix / regex.

It assumes you're using SemVer as your versioning strategy, or at least that your versions are numbers or are prefixed with a custom string, but still end in numbers that are sortable.

For example, let's say you have a custom docker repository running on `192.168.0.100:5000`, and that you have an image `lfourky/awesomeImage` in there. 
If you make a GET request to `192.168.0.100:5000/v2/lfourky/awesomeImage/tags/list`, you'll get the following response: 
```sh
{   
    "name": "lfourky/awesomeImage",
    "tags": [
        "tag2", "tag1", "build_42", "build_32", "build_33", "awesome", "latest", "283"
    ]
}
```
There it is! You spot the tag that you need, and it's `build_42`. But it's not sorted, and you want to retrieve it somehow. Well, that's where LDT steps in.

# Running from source code

```sh
go run ldt.go -h http://192.168.0.100:5000 -n lfourky/awesomeImage -p "build_"
```
This will print out `build_42` to `STDOUT` so you can pipe it somewhere or append it as a result to a command.

# Running from an official docker image
I've built and pushed a docker image to Docker Hub for you to use freely. 
You can build one yourself... or not. Your choice!
```sh
docker run --rm lfourky/last-docker-tag -h "http://192.168.0.100:5000" -n "awesomeImage" -p "build_"
```

# How it works
If you provide a `-p` (tag prefix) flag when you run it, it will find all the tags that start with that specified prefix. Then, it will collect them all, and strip the prefixes so it can sort it using a SemVer library (thanks, HashiCorp!). After that, it will print the latest tag with the prefix prepended to it, or, if you didn't specify a prefix, will just print the latest tag.

## Available flags
| Flag | Description | Type | Default | Example | Required |
| ------------- | ------------- | ------------- | ------------- | ------------- |------------- |
| -h | Docker Repository Host Address | string | / | http://192.168.0.130:5000 | ✔
| -v | Docker Repository Version | string | v2 | / | /
| -n | Image name | string | / | lfourky/awesomeImage |  ✔
| -p | Tag prefix to strip when matching tags | string | / | myCompany_staging_build_ | /
| -l | Verbose logging | boolean | false | true | /

* You can redirect the error to a log file, the default value returned to STDOUT is "latest" if there's an error, e.g.
`go run ldt.go -h http://192.168.0.100:5000 -n lfourky/awesomeImage 2>> errors.log`

# Where I used it
https://github.com/kubernetes/kubernetes/issues/33664
Because of the issue mentioned in this... issue, I'm now using this project as a workaround that works for me and my usecase.
```sh
kubectl set image deployment.v1.apps/notifications notifications=192.168.0.100:5000/ntf/notifications:$(docker run --rm lfourky/last-docker-tag -h http://192.168.0.100:5000 -n ntf/notifications -p build_ 2>>errors.log)
```
