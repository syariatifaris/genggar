## Genggar

Genggar is a simple UDP client - server based library for event sourcing. If you are experimenting a saga pattern, you may be interested to take a look into this repo. At the moment, Genggar is an experimental work. 
 
### Getting Started

Get the latest update of genggar by using go get command, or you can just git clone from our github repository. 

``` 
go get github.com/syariatifaris/genggar
```

We support `go dep` for project dependencies . Run the dep ensure to obtain the latest dependencies for this repo

```
dep ensure -v
```

## How it works

Genggar usage is quite straight forward. The event publisher needs to initiate the genggar server and publish the event message for specific topic, whilst the subscriber listen to the server and accept the event. 

### Server Application

First of all, the server will run on specific address and port. It uses UDP connection to broadcast and receive message from the subscriber. 

```
server, err := genggar.NewEventServer("127.0.0.1", 1234)  
if err != nil {  
   glog.ERROR.Fatalln("server error", err.Error())  
}
```

Then,  the server needs to be started inside a background routine. It will listens for a connection and communicates with the subscribers using a predefined protocol.

```
stopServer := make(chan bool)
go func() {  
   server.Start(stopServer)  
}()
```
To enable the event publishing, the server should execute the event dispatcher. The event dispatcher responsible to send the data inside a event buffer asynchronously to the subscriber with specific topic.
```
stopDispatcher := make(chan bool)   
go func() {  
   server.DispatchEventPublisher(stopDispatcher)  
}()
```

Finally, server will be able to publish any event by using `PublishEvent` function. This function, will read the event to a certain buffer, which will be consumed by the dispatcher. 

```
err := server.PublishEvent("ORDER", "NEW_ORDER_VERIFIED", "publish event message from server")  
  
if err != nil {  
   log.Println("cannot publish event to all", err.Error())  
}
```

The full sample of Genggar server application can be found here [server application](https://github.com/syariatifaris/genggar/tree/master/example/server)

### Subscriber Application

The subscriber will be able to listen for a specific topic, and handle the event inside their predefined callback.  First of all, the subscriber need to connect to the server. 

The subscriber needs to define their event processors, for specific events. It is also possible to bind the events to the multiple event processors. When the event is accepted by the subscriber, it will also run the processor and return the error if the process fail. 

```
func main(){  
   client, err := genggar.NewSubscriberClient(  
      "127.0.0.1", "1234", "ORDER",  
      []*engine.EventProcessor{  
         {  
            Events:   []string{"NEW_ORDER_VERIFIED", "NEW_ORDER_WAITING"},  
            Callback: newOrderEvent,  
         },  
         {  
            Events:   []string{"REJECT_BY_SELLER", "REJECT_BY_SYSTEM"},  
            Callback: rejectOrderEvent,  
         },  
      },  
   )  
}

func newOrderEvent(topic, eventName string, data interface{}) error {  
   glog.INFO.Printf("[processor][topic:%s][event:%s] in processing\n", topic, eventName)  
   return nil  
}  
  
func rejectOrderEvent(topic, eventName string, data interface{}) error {  
   glog.INFO.Printf("[processor][topic:%s][event:%s] in processing\n", topic, eventName)  
   return nil  
}
```

Similarly with the server, the client also needs to listen for the subscriber by running background process for as an active listener

```
go func() {  
   client.StartListen(stopChan)  
}()
``` 

The full sample of Genggar client application can be found here [client application](https://github.com/syariatifaris/genggar/tree/master/example/client)

## To Do(s)

As I mention at the beginning, Genggar is an experimental work. As I know, this work wont pay anyone salary, so it may require lot of time to develop to make it works nicely. 

Nevertheless, I have some point in my mind for the next milestone such as:

 1. Enable retry mechanism when processor return failed
 2. Add an adaptive & smart scheduler
 3. Support multiple database for event histories (I doubt this work wont be called an event sourcing until this feature is implemented) 

## Contribution

 - Fork the transactionapp repository to your remote
 - Add your remote to your local git genggar
 - From **master** branch at your local, checkout to your feature branch, and start the development
 - Make a pull request to staging, and do the test properly after getting approved on our staging server
 - Make a pull request to master

