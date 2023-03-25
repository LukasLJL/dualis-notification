# dualis-notification

This simple Go tool fetches all published grades from Dualis and sends you an email when new grades get published. If you don't know what Dualis is you can definitely call yourself happy. Dualis is the grade management system that is used at the DHBW Stuttgart. This project is one reason never to touch the web interface again.

This project was forked from [github.com/mariuskiessling/dhbw-gradifier](https://github.com/mariuskiessling/dhbw-gradifier), to update some dependencies and add docker support for an easy deployment.

## :rocket: Getting started
Here is an example docker-compose file which you can use 
````yaml
version: "3.8"
services:
  dualis-notification:
    image: ghcr.io/lukasljl/dualis-notification:latest
    container_name: dualis-notification
    restart: unless-stopped
    env_file:
      - ./config.env
````
You can either mount an existing configuration file at `/config.env` or you can simply adjust the environment variables.
In this example the environment variables are provides with the docker feature `env_file`, but you can also use other methods.

