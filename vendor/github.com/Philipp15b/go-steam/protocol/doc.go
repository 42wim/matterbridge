/*
This package includes some basics for the Steam protocol. It defines basic interfaces that are used throughout go-steam:
There is IMsg, which is extended by IClientMsg (sent after logging in) and abstracts over
the outgoing message types. Both interfaces are implemented by ClientMsgProtobuf and ClientMsg.
Msg is like ClientMsg, but it is used for sending messages before logging in.

There is also the concept of a Packet: This is a type for incoming messages where only
the header is deserialized. It therefore only contains EMsg data, job information and the remaining data.
Its contents can then be read via the Read* methods which read data into a MessageBody - a type which is Serializable and
has an EMsg.

In addition, there are extra types for communication with the Game Coordinator (GC) included in the gamecoordinator sub-package.
For outgoing messages the IGCMsg interface is used which is implemented by GCMsgProtobuf and GCMsg.
Incoming messages are of the GCPacket type and are read like regular Packets.

The actual messages and enums are in the sub-packages steamlang and protobuf, generated from the SteamKit data.
*/
package protocol
