# List of Docker images for `act`

**Warning:** Below badges with size for each image are displaying size of **compressed image size in registry. After pulling the image, size can be drastically different due to Docker uncompressing the image layers.**

## Images based on [`buildpack-deps`][hub/_/buildpack-deps]

**Note 1: `node` images are based on Debian root filesystem, while it is extremely similar to Ubuntu, there might be some differences**

**Note 2: `node` `-slim` images don't have `python` installed, if you want to use actions or software that is depending on `python`, you need to specify image manually**

**Note 3: Default micro platforms shipped with act use current Node LTS slim tags (Node 24 / Node 20). Prefer `catthehacker/ubuntu:act-*` for broader action compatibility.**

| Image                                   | Size                                                         |
| --------------------------------------- | ------------------------------------------------------------ |
| [`node:22-bookworm`][hub/_/node]        | ![`bookworm24-size`][hub/_/node/24-bookworm/size]            |
| [`node:22-bookworm-slim`][hub/_/node]   | ![`micro-bookworm24-size`][hub/_/node/24-bookworm-slim/size] |
| [`node:20-bookworm`][hub/_/node]        | ![`bookworm20-size`][hub/_/node/20-bookworm/size]            |
| [`node:20-bookworm-slim`][hub/_/node]   | ![`micro-bookworm20-size`][hub/_/node/20-bookworm-slim/size] |
| [`node:20-bullseye`][hub/_/node]        | ![`bullseye20-size`][hub/_/node/20-bullseye/size]            |
| [`node:20-bullseye-slim`][hub/_/node]   | ![`micro-bullseye20-size`][hub/_/node/20-bullseye-slim/size] |

**Note: `catthehacker/ubuntu` images are based on Ubuntu root filesystem**

| Image                                                        | GitHub Repository                                             |
| ------------------------------------------------------------ | ------------------------------------------------------------- |
| [`catthehacker/ubuntu:act-latest`][ghcr/catthehacker/ubuntu] | [`catthehacker/docker-images`][gh/catthehacker/docker_images] |
| [`catthehacker/ubuntu:act-24.04`][ghcr/catthehacker/ubuntu]  | [`catthehacker/docker-images`][gh/catthehacker/docker_images] |
| [`catthehacker/ubuntu:act-22.04`][ghcr/catthehacker/ubuntu]  | [`catthehacker/docker-images`][gh/catthehacker/docker_images] |
| [`catthehacker/ubuntu:act-20.04`][ghcr/catthehacker/ubuntu]  | [`catthehacker/docker-images`][gh/catthehacker/docker_images] |

## Images based on [`actions/virtual-environments`][gh/actions/virtual-environments]

**Note: `nektos/act-environments-ubuntu` have been last updated in February, 2020. It's recommended to update the image manually after `docker pull` if you decide to use it. Prefer `catthehacker/ubuntu:full-*` instead.**

| Image                                                                             | Size                                                                       | GitHub Repository                                       |
| --------------------------------------------------------------------------------- | -------------------------------------------------------------------------- | ------------------------------------------------------- |
| [`nektos/act-environments-ubuntu:18.04`][hub/nektos/act-environments-ubuntu]      | ![`nektos:18.04`][hub/nektos/act-environments-ubuntu/18.04/size]           | [`nektos/act-environments`][gh/nektos/act-environments] |
| [`nektos/act-environments-ubuntu:18.04-lite`][hub/nektos/act-environments-ubuntu] | ![`nektos:18.04-lite`][hub/nektos/act-environments-ubuntu/18.04-lite/size] | [`nektos/act-environments`][gh/nektos/act-environments] |
| [`nektos/act-environments-ubuntu:18.04-full`][hub/nektos/act-environments-ubuntu] | ![`nektos:18.04-full`][hub/nektos/act-environments-ubuntu/18.04-full/size] | [`nektos/act-environments`][gh/nektos/act-environments] |

| Image                                                         | GitHub Repository                                                                     |
| ------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| [`catthehacker/ubuntu:full-latest`][ghcr/catthehacker/ubuntu] | [`catthehacker/virtual-environments-fork`][gh/catthehacker/virtual-environments-fork] |
| [`catthehacker/ubuntu:full-22.04`][ghcr/catthehacker/ubuntu]  | [`catthehacker/virtual-environments-fork`][gh/catthehacker/virtual-environments-fork] |
| [`catthehacker/ubuntu:full-20.04`][ghcr/catthehacker/ubuntu]  | [`catthehacker/virtual-environments-fork`][gh/catthehacker/virtual-environments-fork] |

Feel free to make a pull request with your image added here

[hub/_/buildpack-deps]: https://hub.docker.com/_/buildpack-deps
[hub/_/node]: https://hub.docker.com/r/_/node
[hub/_/node/24-bookworm/size]: https://img.shields.io/docker/image-size/_/node/24-bookworm
[hub/_/node/24-bookworm-slim/size]: https://img.shields.io/docker/image-size/_/node/24-bookworm-slim
[hub/_/node/20-bookworm/size]: https://img.shields.io/docker/image-size/_/node/20-bookworm
[hub/_/node/20-bookworm-slim/size]: https://img.shields.io/docker/image-size/_/node/20-bookworm-slim
[hub/_/node/20-bullseye/size]: https://img.shields.io/docker/image-size/_/node/20-bullseye
[hub/_/node/20-bullseye-slim/size]: https://img.shields.io/docker/image-size/_/node/20-bullseye-slim
[ghcr/catthehacker/ubuntu]: https://github.com/catthehacker/docker_images/pkgs/container/ubuntu
[hub/nektos/act-environments-ubuntu]: https://hub.docker.com/r/nektos/act-environments-ubuntu
[hub/nektos/act-environments-ubuntu/18.04/size]: https://img.shields.io/docker/image-size/nektos/act-environments-ubuntu/18.04
[hub/nektos/act-environments-ubuntu/18.04-lite/size]: https://img.shields.io/docker/image-size/nektos/act-environments-ubuntu/18.04-lite
[hub/nektos/act-environments-ubuntu/18.04-full/size]: https://img.shields.io/docker/image-size/nektos/act-environments-ubuntu/18.04-full

<!-- GitHub repository links -->

[gh/nektos/act-environments]: https://github.com/nektos/act-environments
[gh/actions/virtual-environments]: https://github.com/actions/virtual-environments
[gh/catthehacker/docker_images]: https://github.com/catthehacker/docker_images
[gh/catthehacker/virtual-environments-fork]: https://github.com/catthehacker/virtual-environments-fork
