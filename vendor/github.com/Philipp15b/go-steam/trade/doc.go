/*
Allows automation of Steam Trading.

Usage

Like go-steam, this package is event-based. Call Poll() until the trade has ended, that is until the TradeEndedEvent is emitted.

	// After receiving the steam.TradeSessionStartEvent
	t := trade.New(sessionIdCookie, steamLoginCookie, steamLoginSecure, event.Other)
	for {
		eventList, err := t.Poll()
		if err != nil {
			// error handling here
			continue
		}
		for _, event := range eventList {
			switch e := event.(type) {
				case *trade.ChatEvent:
					// respond to any chat message
					t.Chat("Trading is awesome!")
				case *trade.TradeEndedEvent:
					return
				// other event handlers here
			}
		}
	}

You can either log into steamcommunity.com and use the values of the `sessionId` and `steamLogin` cookies,
or use go-steam and after logging in with client.Web.LogOn() and receiving the WebLoggedOnEvent use the `SessionId`
and `SteamLogin` fields of steam.Web for the respective cookies.

It is important that there is no delay between the Poll() calls greater than the timeout of the Steam client
(currently five seconds before the trade partner sees a warning) or the trade will be closed automatically by Steam.

Notes

All method calls to Steam APIs are blocking. This packages' and its subpackages' types are not thread-safe and no calls to any method of the same
trade instance may be done concurrently except when otherwise noted.
*/
package trade
