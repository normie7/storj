# storj homework problem
solution to [https://gist.githubusercontent.com/jtolds/0cde4aa3e07b20d6a42686ad3bc9cb53](https://gist.githubusercontent.com/jtolds/0cde4aa3e07b20d6a42686ad3bc9cb53)

### choosing protocol
What are our options when choosing protocol for this task? 

We will build our protocol over TCP.

We could separate the data we will be sending into two categories:
 - some service information
 - file data
 
We could use messages of fixed size or type-length-value messages. With fixed size messages
 it will be easier for us to encode and decode messages, but since most of our data will come 
 from file and not intended to be humanly readable, we will get huge overhead of splitting 
 the file into small pieces and attaching header for each piece for nothing. So I decided to use
 type-length-value system instead. 
 
With TLV we could encode our service information messages any way we like (gob or json or anything else).
But as it was enough for this particular project to just send a string as a value. 

While file data intended to be saved directly onto disk, service messages are intended to be stored in memory, 
that's why I decided to limit max message size.

The protocol is stateful. But since we basically got only 1 message per state we don't have to store it.

### Installation
`go install ./...`

### Scenario

Relay is waiting for 2 types of client: sender and receiver. 

Sender sends registration message and waits for reply. If everything is ok, sender will print secret code and wait for StartFileTransferCommand. 
Relay will save sender connection and it's secret code. 

Receiver sends registration message with secret code. If sender is found Relay will match the client and enter the "proxy mode".

If registration went ok, receiver will send StarFileTransfer command.

Sender will send a message with filename in it. Then it will send file content.

If everything is ok both clients will close the connection.

### what could be changed for production
 
 - add relay graceful shutdown
 - proper timeouts for connections. current implementation is open for attack scenario 
 if we just open the connection and don't write anything in it.
 - proper errors for clients. now if anything goes wrong they just close the connection.
 - proper logging. we can't put anything in stdout except errors in the current implementation because of the requirements.
 - tests
     

 