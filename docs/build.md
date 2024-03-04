## Compilation

Steps:

1. Install Go >= 1.21.
2. Get the source code: `git clone https://github.com/jgkawell/yarr.git`
3. Build: `go build -o bin/yarr .`

Using Docker:

```bash
docker run --rm -u $(id -u):$(id -g) -v "$PWD":$PWD -w $PWD -e GOCACHE=$PWD/.cache golang go build -o bin/yarr
```
