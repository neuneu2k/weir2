# weir2
Hyena Http to messaging gateway

**Warning: Alpha quality code, not for production use**

## About

Weir2 exposes Hyena Infrastructure services to the outside world, it's role is to
* Terminate SSL
* Translate HTTP (1.1 and 2.0) to internal streaming RPCs
* Forward to the hyena daemon for internal routing 
* Serve static content
* Proxy user message queues
 
## Current status

- [X] Http 1.1 listener
- [ ] Https 1.1 listener (untested)
- [X] Protocol conversion
- [ ] Metrics
- [ ] Metrics publication
- [ ] Static content serving
- [ ] User queue proxying
- [ ] Quality GoDoc comments


