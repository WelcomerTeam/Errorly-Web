echo "Build GO Executable"
go build -v -o errorly cmd/main.go

echo "Build Web Distributable"
#!cd web
#!yarn lint
#!yarn build --modern
#!cd ..

echo "Docker build and push"
docker build --tag 1345/errorly:latest .
docker push 1345/errorly:latest
