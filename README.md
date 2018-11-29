# LDT

LDT retrieves the latest tag, out of list of tags, retrieved from a custom docker repository, while matching the tag by prefix / regex.

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

You can also use a regular expression to target it:
```sh
go run ldt.go -h http://192.168.0.100:5000 -n lfourky/awesomeImage -r "^build_\d+$"
```
This will produce the same result.

# Running from an official docker image
I've built and pushed a docker image to Docker Hub for you to use freely. 
You can build one yourself... or not. Your choice!
```sh
docker run --rm lfourky/last-docker-tag -h "http://192.168.0.100:5000" -n "awesomeImage" -p "build_"
```

## Available flags
```sh
-v Version (string)
   Docker repository version (defaults to "v2")
-h Host (string)
   Full docker repository address. e.g. "http://192.168.0.130:5000"
-n Repository name (string)
   Full repository name. e.g. "library/mysql"
-p Target tag prefix (string)
   Prefix that you want to use for string matching. e.g. "build_"
-r Target tag regex (string)
   Regex that you want to use for string matching. e.g. "^build_\d+$"
```

* Do not use both prefix and regex at the same time, pick one.

# Where I used it
https://github.com/kubernetes/kubernetes/issues/33664
So I've started using Kubernetes and Drone CI for myself lately. I've found that, when I publish an image to my docker repository, and then try to apply a newer version with `kubectl set image...` and the tag is `latest`, even though there's a newer version in my docker repo, the image is not being reapplied / repulled. Thus, a workaround (this project) I'm now using is:
```sh
kubectl set image deployment.v1.apps/notifications notifications=192.168.0.100:5000/ntf/notifications:$(docker run --rm lfourky/last-docker-tag -h http://192.168.0.100:5000 -n ntf/notifications -p build_)
```
