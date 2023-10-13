TAG="europe-north1-docker.pkg.dev/spherical-realm-401810/good-blast/api"

docker build -t ${TAG} .

docker push ${TAG}