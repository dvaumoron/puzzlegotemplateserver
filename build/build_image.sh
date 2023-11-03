#!/usr/bin/env bash

./build/build.sh

buildah from --name puzzlegotemplateserver-working-container scratch
buildah copy puzzlegotemplateserver-working-container $HOME/go/bin/puzzlegotemplateserver /bin/puzzlegotemplateserver
buildah copy puzzlegotemplateserver-working-container ../puzzletest/templatedata/templates /templates
buildah copy puzzlegotemplateserver-working-container ../puzzletest/templatedata/locales /locales
buildah config --env COMPONENTS_PATH=templates/components puzzlegotemplateserver-working-container
buildah config --env VIEWS_PATH=templates/views puzzlegotemplateserver-working-container
buildah config --env LOCALES_PATH=/locales puzzlegotemplateserver-working-container
buildah config --env SERVICE_PORT=50051 puzzlegotemplateserver-working-container
buildah config --port 50051 puzzlegotemplateserver-working-container
buildah config --entrypoint '["/bin/puzzlegotemplateserver"]' puzzlegotemplateserver-working-container
buildah commit puzzlegotemplateserver-working-container puzzlegotemplateserver
buildah rm puzzlegotemplateserver-working-container

buildah push puzzlegotemplateserver docker-daemon:puzzlegotemplateserver:latest
