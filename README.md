# Amalgam8 Room for Game On!
![Game On! Logo](https://avatars3.githubusercontent.com/u/15149525?v=3&s=200) ![Amalgam8 Logo](https://avatars2.githubusercontent.com/u/19418902?v=3&s=200)  

## What is Game On!?

*Game On!* is a text-based adventure game built to explore and demonstrate microservices architecture and related concepts.
The game is extensible by allowing custom game rooms to be added to it.  
You can learn more about *Game On!* by visiting its [official website](https://game-on.org).

## What is Amalgam8?

*Amalgam8* is a content and version-based routing fabric for polyglot microservices.
It provides a centralized control plane which can be used to dynamically program rules for routing and manipulating requests across microservices in a running application.
It also provides a sidecar which automates service discovery, registration, and request routing.  
You can learn more about *Amalgam8* by visiting its [official website](https://www.amalgam8.io/), or its [GitHub repository](https://github.com/amalgam8/amalgam8).

## What is the Amalgam8 Room?

The *Amalgam8 Room* is a custom *Game On!* room, built to demonstrate the *Amalgam8* framework and its capabilities.

The *Amalgam8 Room* is built as a microservices application. It's composed of two distinct microservices:
- The *room* service.  
  This service won't make your bed sheets, but it implements the core functionality of the room over a simple REST API.  
  It uses the *Amalgam8* framework to automate service registration.
- The *mediator* service.  
  This service is used as the front-facing service for the room. It's responsible for managing the websocket connections with *Game On!*'s own [mediator service](https://gameontext.gitbooks.io/gameon-gitbook/content/microservices/#_mediator),
  as well as forwarding incoming [game commands](https://gameontext.gitbooks.io/gameon-gitbook/content/microservices/WebSocketProtocol.html), over HTTP, to the *room* service.  
  It uses the *Amalgam8* framework for discoverying *room* service endpoints, and proxying its calls to it.

## Prerequisites

To build and run *Amalgam8 Room*, you'll need *Go* 1.6+, *docker* 1.10+, *docker-compose* 1.51+, and the [Amalgam8 CLI](https://github.com/amalgam8/a8ctl).

## Run the room locally

1. Obtain the source of this repository:  
  `go get github.com/gameontext/a8-roo`
2. Navigate to the Amalgam8 Room root directory:  
 `cd $GOPATH/src/github.com/gameontext/a8-room`
3. Build the game binaries and docker images:  
  `make build dockerize`
4. Run the room:  
  `make start`

You can use `docker ps` to verify that the room's containers are running. Expect to see Amalgam8's controlplane services (*a8room_controller*, *a8room_registry* and *a8room_redis*), as well as the room's own *a8room_room* and *a8room_mediator* microservices.

## Register the room with Game On!

1. Enter the Game On! website with your web browser:
   - If you're [running Game On! locally](https://github.com/gameontext/gameon#local-room-development), the website is available at http://127.0.0.1.  
   - You can also use the [hosted version](https://game-on.org) of the game, but you'll have to make sure your room service is publicly reachable (e.g., by deploying it to [Bluemix](https://console.ng.bluemix.net/)).

2. Login to the game. If you're running Game On! locally, you can only login as an anonymous user.

3. Click the "Edit rooms" button, at the top-right of the screen.

4. Regenerate a shared secret, and select "Create a new room" from the dropdown menu.

5. Set a room ID and name. Use "Amalgam8" for both.

6. Set the websocket endpoint for the room:  
   - If you're running locally, the room is available at [ws://&lt;docker0_ip&gt;:3000](ws://&lt;docker0_ip&gt;:3000).
     To get the IP address of the docker0 interface, use the following command:
     ```shell
     ifconfig docker0 | grep 'inet addr:' | cut -d: -f2 | awk '{ print $1}'
     ```
     (Note: the room is also available at [ws://127.0.0.1:3000](ws://127.0.0.1:3000), but we need to use an address reachable from within the docker containers in which the Game On! services are running).
   - If you're using the hosted version of Game On!, make sure you assign a publicly reachable address to the room's mediator service, and expose port 3000.

## Use the room

The easiest way to get to the room is by going to the first room (`/sos`), and teleporting to the Amalgam8 room (`/teleport Amalgam8`).
Once you're in, you can examine the room, or just chat around.

## Deploying a new version of the room

To demonstrate content and based-version routing with Amalgam8, we'll deploy a new version of the room service.  
The new version will be deployed into a live system, and we'll use Amalgam8 to only expose it to a select set of players, so that we can test and verify its behavior without impacting other players.

1. Set the Amalgam8 controlplane endpoints for the [Amalgam8 CLI](https://github.com/amalgam8/a8ctl):   
   ```shell
   export A8_REGISTRY_URL=http://127.0.0.1:8001
   export A8_CONTROLLER_URL=http://127.0.0.1:8002
   ```

2. Set a routing rule such that traffic will be routed by default to the old version ("v1") of the room service:  
   ```shell
   a8ctl route-set --default v1 room
   ```
   This allows us to deploy new version of the room microservice, without routing traffic to it just yet.
   
3. Deploy the new version ("v2") of the room service:
   ```shell
   docker run -d --name a8room_room_2 \
     --env VERSION=v2 \
     --env A8_SERVICE=room:v2 \
     --net a8room_default \
     --link registry --link controller \
     gameon-a8-room/room:latest
   ```
   For simplicitly, the docker image for the room service already support both the "v1" and "v2" versions of it.
   We set `VERSION=v2` to enable it. We also set `A8_SERVICE=room:v2` to let Amalgam8 know this is the "v2" version of the room service.  
   Note: we make sure to run the container on the same `a8room_default` network create by docker-compose.

4. Set a routing rule such that traffic will be routed to room service v2 for our test player only:
   ```shell
   a8ctl route-set --source mediator --default v1 --selector "v2(header=X-Game-On-Username:GiantMuffin)" room
   ```
   (Make sure to replace "GiantMuffin" with your own username).  
   Here, we take advantage of the fact that the mediator service, when calling the room service over its REST API, adds the `X-Game-On-User-Id`, and `X-Game-On-Username` headers, indicating the player associated with the command.
 
5. The new version of the room service includes a built-in profanity checker, preventing playes from swearing in the room.  
   We can test it by entering the room as the "GiantMuffin" test player, and start swearing around! (note: make sure not to get too rude... "poop" or "boogers" will make due).  
   Note that players other than "GiantMuffin" will not be exposed to the functionality provided by the new version of the room service.
   
6. Once we're confident that the new version is stable, we can expose it to the rest of the players:
    ```shell
    a8ctl route-set --default v2 room
    ```
    
## What to do next

Checkout [Amalgam8's demo apps](https://www.amalgam8.io/docs/demo.html) for some other stuff you can do with Amalgam8.

## Cleanup
To tear down the Amalgam8 room, just use the following:
```shell
make stop
```
