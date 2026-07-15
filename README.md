# chat behind your NAT with a peer using `p2punch`

p2punch uses a technique called [UDP hole punching](https://en.wikipedia.org/wiki/UDP_hole_punching) for establishing bidirectional UDP connections between internet hosts

## usage
deploy the server first:
```console
docker build -f Containerfile -t p2punch .
docker run -d --name p2punch -p 6969:6969/udp p2punch

# make sure your firewall has the port open for udp
sudo ufw allow 6969/udp
```

and then connect with a client, specifying the room name for you and your peer:
```console
go build ./cmd/client
./client --room huckleberry --addr 93.184.216.34
```
> **warning:** do not specify the port for the `addr` flag. p2punch prefers 6969, please respect it

