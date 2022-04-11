// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package girc

// Standard CTCP based constants.
const (
	CTCP_ACTION     = "ACTION"
	CTCP_PING       = "PING"
	CTCP_PONG       = "PONG"
	CTCP_VERSION    = "VERSION"
	CTCP_USERINFO   = "USERINFO"
	CTCP_CLIENTINFO = "CLIENTINFO"
	CTCP_SOURCE     = "SOURCE"
	CTCP_TIME       = "TIME"
	CTCP_FINGER     = "FINGER"
	CTCP_ERRMSG     = "ERRMSG"
)

// Emulated event commands used to allow easier hooks into the changing
// state of the client.
const (
	UPDATE_STATE     = "CLIENT_STATE_UPDATED"   // when channel/user state is updated.
	UPDATE_GENERAL   = "CLIENT_GENERAL_UPDATED" // when general state (client nick, server name, etc) is updated.
	ALL_EVENTS       = "*"                      // trigger on all events
	CONNECTED        = "CLIENT_CONNECTED"       // when it's safe to send arbitrary commands (joins, list, who, etc), trailing is host:port
	INITIALIZED      = "CLIENT_INIT"            // verifies successful socket connection, trailing is host:port
	DISCONNECTED     = "CLIENT_DISCONNECTED"    // occurs when we're disconnected from the server (user-requested or not)
	CLOSED           = "CLIENT_CLOSED"          // occurs when Client.Close() has been called
	STS_UPGRADE_INIT = "STS_UPGRADE_INIT"       // when an STS upgrade initially happens.
	STS_ERR_FALLBACK = "STS_ERR_FALLBACK"       // when an STS connection fails and fallbacks are supported.
)

// User/channel prefixes :: RFC1459.
const (
	DefaultPrefixes = "(ov)@+" // the most common default prefixes
	ModeAddPrefix   = "+"      // modes are being added
	ModeDelPrefix   = "-"      // modes are being removed

	ChannelPrefix      = "#" // regular channel
	DistributedPrefix  = "&" // distributed channel
	OwnerPrefix        = "~" // user owner +q (non-rfc)
	AdminPrefix        = "&" // user admin +a (non-rfc)
	HalfOperatorPrefix = "%" // user half operator +h (non-rfc)
	OperatorPrefix     = "@" // user operator +o
	VoicePrefix        = "+" // user has voice +v
)

// User modes :: RFC1459; section 4.2.3.2.
const (
	UserModeInvisible     = "i" // invisible
	UserModeOperator      = "o" // server operator
	UserModeServerNotices = "s" // user wants to receive server notices
	UserModeWallops       = "w" // user wants to receive wallops
)

// Channel modes :: RFC1459; section 4.2.3.1.
const (
	ModeDefaults = "beI,k,l,imnpst" // the most common default modes

	ModeInviteOnly = "i" // only join with an invite
	ModeKey        = "k" // channel password
	ModeLimit      = "l" // user limit
	ModeModerated  = "m" // only voiced users and operators can talk
	ModeOperator   = "o" // operator
	ModePrivate    = "p" // private
	ModeSecret     = "s" // secret
	ModeTopic      = "t" // must be op to set topic
	ModeVoice      = "v" // speak during moderation mode

	ModeOwner        = "q" // owner privileges (non-rfc)
	ModeAdmin        = "a" // admin privileges (non-rfc)
	ModeHalfOperator = "h" // half-operator privileges (non-rfc)
)

// IRC commands :: RFC2812; section 3 :: RFC2813; section 4.
const (
	ADMIN    = "ADMIN"
	AWAY     = "AWAY"
	CONNECT  = "CONNECT"
	DIE      = "DIE"
	ERROR    = "ERROR"
	INFO     = "INFO"
	INVITE   = "INVITE"
	ISON     = "ISON"
	JOIN     = "JOIN"
	KICK     = "KICK"
	KILL     = "KILL"
	LINKS    = "LINKS"
	LIST     = "LIST"
	LUSERS   = "LUSERS"
	MODE     = "MODE"
	MOTD     = "MOTD"
	NAMES    = "NAMES"
	NICK     = "NICK"
	NJOIN    = "NJOIN"
	NOTICE   = "NOTICE"
	OPER     = "OPER"
	PART     = "PART"
	PASS     = "PASS"
	PING     = "PING"
	PONG     = "PONG"
	PRIVMSG  = "PRIVMSG"
	QUIT     = "QUIT"
	REHASH   = "REHASH"
	RESTART  = "RESTART"
	SERVER   = "SERVER"
	SERVICE  = "SERVICE"
	SERVLIST = "SERVLIST"
	SQUERY   = "SQUERY"
	SQUIT    = "SQUIT"
	STATS    = "STATS"
	SUMMON   = "SUMMON"
	TIME     = "TIME"
	TOPIC    = "TOPIC"
	TRACE    = "TRACE"
	USER     = "USER"
	USERHOST = "USERHOST"
	USERS    = "USERS"
	VERSION  = "VERSION"
	WALLOPS  = "WALLOPS"
	WEBIRC   = "WEBIRC"
	WHO      = "WHO"
	WHOIS    = "WHOIS"
	WHOWAS   = "WHOWAS"
)

// Numeric IRC reply mapping :: RFC2812; section 5.
const (
	RPL_WELCOME           = "001"
	RPL_YOURHOST          = "002"
	RPL_CREATED           = "003"
	RPL_MYINFO            = "004"
	RPL_BOUNCE            = "005"
	RPL_ISUPPORT          = "005"
	RPL_USERHOST          = "302"
	RPL_ISON              = "303"
	RPL_AWAY              = "301"
	RPL_UNAWAY            = "305"
	RPL_NOWAWAY           = "306"
	RPL_WHOISUSER         = "311"
	RPL_WHOISSERVER       = "312"
	RPL_WHOISOPERATOR     = "313"
	RPL_WHOISIDLE         = "317"
	RPL_ENDOFWHOIS        = "318"
	RPL_WHOISCHANNELS     = "319"
	RPL_WHOWASUSER        = "314"
	RPL_ENDOFWHOWAS       = "369"
	RPL_LISTSTART         = "321"
	RPL_LIST              = "322"
	RPL_LISTEND           = "323" //nolint:misspell // it's correct.
	RPL_UNIQOPIS          = "325"
	RPL_CHANNELMODEIS     = "324"
	RPL_NOTOPIC           = "331"
	RPL_TOPIC             = "332"
	RPL_INVITING          = "341"
	RPL_SUMMONING         = "342"
	RPL_INVITELIST        = "346"
	RPL_ENDOFINVITELIST   = "347"
	RPL_EXCEPTLIST        = "348"
	RPL_ENDOFEXCEPTLIST   = "349"
	RPL_VERSION           = "351"
	RPL_WHOREPLY          = "352"
	RPL_ENDOFWHO          = "315"
	RPL_NAMREPLY          = "353"
	RPL_ENDOFNAMES        = "366"
	RPL_LINKS             = "364"
	RPL_ENDOFLINKS        = "365"
	RPL_BANLIST           = "367"
	RPL_ENDOFBANLIST      = "368"
	RPL_INFO              = "371"
	RPL_ENDOFINFO         = "374"
	RPL_MOTDSTART         = "375"
	RPL_MOTD              = "372"
	RPL_ENDOFMOTD         = "376"
	RPL_YOUREOPER         = "381"
	RPL_REHASHING         = "382"
	RPL_YOURESERVICE      = "383"
	RPL_TIME              = "391"
	RPL_USERSSTART        = "392"
	RPL_USERS             = "393"
	RPL_ENDOFUSERS        = "394"
	RPL_NOUSERS           = "395"
	RPL_TRACELINK         = "200"
	RPL_TRACECONNECTING   = "201"
	RPL_TRACEHANDSHAKE    = "202"
	RPL_TRACEUNKNOWN      = "203"
	RPL_TRACEOPERATOR     = "204"
	RPL_TRACEUSER         = "205"
	RPL_TRACESERVER       = "206"
	RPL_TRACESERVICE      = "207"
	RPL_TRACENEWTYPE      = "208"
	RPL_TRACECLASS        = "209"
	RPL_TRACERECONNECT    = "210"
	RPL_TRACELOG          = "261"
	RPL_TRACEEND          = "262"
	RPL_STATSLINKINFO     = "211"
	RPL_STATSCOMMANDS     = "212"
	RPL_ENDOFSTATS        = "219"
	RPL_STATSUPTIME       = "242"
	RPL_STATSOLINE        = "243"
	RPL_UMODEIS           = "221"
	RPL_SERVLIST          = "234"
	RPL_SERVLISTEND       = "235"
	RPL_LUSERCLIENT       = "251"
	RPL_LUSEROP           = "252"
	RPL_LUSERUNKNOWN      = "253"
	RPL_LUSERCHANNELS     = "254"
	RPL_LUSERME           = "255"
	RPL_ADMINME           = "256"
	RPL_ADMINLOC1         = "257"
	RPL_ADMINLOC2         = "258"
	RPL_ADMINEMAIL        = "259"
	RPL_TRYAGAIN          = "263"
	ERR_NOSUCHNICK        = "401"
	ERR_NOSUCHSERVER      = "402"
	ERR_NOSUCHCHANNEL     = "403"
	ERR_CANNOTSENDTOCHAN  = "404"
	ERR_TOOMANYCHANNELS   = "405"
	ERR_WASNOSUCHNICK     = "406"
	ERR_TOOMANYTARGETS    = "407"
	ERR_NOSUCHSERVICE     = "408"
	ERR_NOORIGIN          = "409"
	ERR_NORECIPIENT       = "411"
	ERR_NOTEXTTOSEND      = "412"
	ERR_NOTOPLEVEL        = "413"
	ERR_WILDTOPLEVEL      = "414"
	ERR_BADMASK           = "415"
	ERR_INPUTTOOLONG      = "417"
	ERR_UNKNOWNCOMMAND    = "421"
	ERR_NOMOTD            = "422"
	ERR_NOADMININFO       = "423"
	ERR_FILEERROR         = "424"
	ERR_NONICKNAMEGIVEN   = "431"
	ERR_ERRONEUSNICKNAME  = "432"
	ERR_NICKNAMEINUSE     = "433"
	ERR_NICKCOLLISION     = "436"
	ERR_UNAVAILRESOURCE   = "437"
	ERR_USERNOTINCHANNEL  = "441"
	ERR_NOTONCHANNEL      = "442"
	ERR_USERONCHANNEL     = "443"
	ERR_NOLOGIN           = "444"
	ERR_SUMMONDISABLED    = "445"
	ERR_USERSDISABLED     = "446"
	ERR_NOTREGISTERED     = "451"
	ERR_NEEDMOREPARAMS    = "461"
	ERR_ALREADYREGISTRED  = "462"
	ERR_NOPERMFORHOST     = "463"
	ERR_PASSWDMISMATCH    = "464"
	ERR_YOUREBANNEDCREEP  = "465"
	ERR_YOUWILLBEBANNED   = "466"
	ERR_KEYSET            = "467"
	ERR_CHANNELISFULL     = "471"
	ERR_UNKNOWNMODE       = "472"
	ERR_INVITEONLYCHAN    = "473"
	ERR_BANNEDFROMCHAN    = "474"
	ERR_BADCHANNELKEY     = "475"
	ERR_BADCHANMASK       = "476"
	ERR_NOCHANMODES       = "477"
	ERR_BANLISTFULL       = "478"
	ERR_NOPRIVILEGES      = "481"
	ERR_CHANOPRIVSNEEDED  = "482"
	ERR_CANTKILLSERVER    = "483"
	ERR_RESTRICTED        = "484"
	ERR_UNIQOPPRIVSNEEDED = "485"
	ERR_NOOPERHOST        = "491"
	ERR_UMODEUNKNOWNFLAG  = "501"
	ERR_USERSDONTMATCH    = "502"
)

// IRCv3 commands and extensions :: http://ircv3.net/irc/.
const (
	AUTHENTICATE = "AUTHENTICATE"
	MONITOR      = "MONITOR"
	STARTTLS     = "STARTTLS"

	CAP       = "CAP"
	CAP_ACK   = "ACK"
	CAP_CLEAR = "CLEAR"
	CAP_END   = "END"
	CAP_LIST  = "LIST"
	CAP_LS    = "LS"
	CAP_NAK   = "NAK"
	CAP_REQ   = "REQ"
	CAP_NEW   = "NEW"
	CAP_DEL   = "DEL"

	CAP_CHGHOST = "CHGHOST"
	CAP_AWAY    = "AWAY"
	CAP_ACCOUNT = "ACCOUNT"
	CAP_TAGMSG  = "TAGMSG"
)

// Numeric IRC reply mapping for ircv3 :: http://ircv3.net/irc/.
const (
	RPL_LOGGEDIN     = "900"
	RPL_LOGGEDOUT    = "901"
	RPL_NICKLOCKED   = "902"
	RPL_SASLSUCCESS  = "903"
	ERR_SASLFAIL     = "904"
	ERR_SASLTOOLONG  = "905"
	ERR_SASLABORTED  = "906"
	ERR_SASLALREADY  = "907"
	RPL_SASLMECHS    = "908"
	RPL_STARTTLS     = "670"
	ERR_STARTTLS     = "691"
	RPL_MONONLINE    = "730"
	RPL_MONOFFLINE   = "731"
	RPL_MONLIST      = "732"
	RPL_ENDOFMONLIST = "733"
	ERR_MONLISTFULL  = "734"
)

// Numeric IRC event mapping :: RFC2812; section 5.3.
const (
	RPL_STATSCLINE    = "213"
	RPL_STATSNLINE    = "214"
	RPL_STATSILINE    = "215"
	RPL_STATSKLINE    = "216"
	RPL_STATSQLINE    = "217"
	RPL_STATSYLINE    = "218"
	RPL_SERVICEINFO   = "231"
	RPL_ENDOFSERVICES = "232"
	RPL_SERVICE       = "233"
	RPL_STATSVLINE    = "240"
	RPL_STATSLLINE    = "241"
	RPL_STATSHLINE    = "244"
	RPL_STATSSLINE    = "245"
	RPL_STATSPING     = "246"
	RPL_STATSBLINE    = "247"
	RPL_STATSDLINE    = "250"
	RPL_NONE          = "300"
	RPL_WHOISCHANOP   = "316"
	RPL_KILLDONE      = "361"
	RPL_CLOSING       = "362"
	RPL_CLOSEEND      = "363"
	RPL_INFOSTART     = "373"
	RPL_MYPORTIS      = "384"
	ERR_NOSERVICEHOST = "492"
)

// Misc.
const (
	ERR_TOOMANYMATCHES = "416" // IRCNet.
	RPL_GLOBALUSERS    = "266" // aircd/hybrid/bahamut, used on freenode.
	RPL_LOCALUSERS     = "265" // aircd/hybrid/bahamut, used on freenode.
	RPL_TOPICWHOTIME   = "333" // ircu, used on freenode.
	RPL_WHOSPCRPL      = "354" // ircu, used on networks with WHOX support.
	RPL_CREATIONTIME   = "329"
)
