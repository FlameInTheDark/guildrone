package guildrone

var (
	APIVersion = "1"

	EndpointGuildedWebsocket = "wss://api.guilded.gg/v1/websocket"

	EndpointGuilded  = "https://www.guilded.gg/"
	EndpointAPI      = EndpointGuilded + "api/v" + APIVersion + "/"
	EndpointChannels = EndpointAPI + "channels/"
	EndpointServers  = EndpointAPI + "servers/"
	EndpointGroups   = EndpointAPI + "groups/"

	EndpointChannel              = func(cID string) string { return EndpointChannels + cID }
	EndpointServer               = func(sID string) string { return EndpointServers + sID }
	EndpointChannelMessages      = func(cID string) string { return EndpointChannels + cID + "/messages" }
	EndpointChannelMessage       = func(cID, mID string) string { return EndpointChannels + cID + "/messages/" + mID }
	EndpointServerMembers        = func(sID string) string { return EndpointServers + sID + "/members" }
	EndpointServerMember         = func(sID, uID string) string { return EndpointServers + sID + "/members/" + uID }
	EndpointServerMemberNickname = func(sID, uID string) string { return EndpointServers + sID + "/members/" + uID + "/nickname" }
	EndpointServerBans           = func(sID string) string { return EndpointServers + sID + "/bans" }
	EndpointServerBansMember     = func(sID, uID string) string { return EndpointServers + sID + "/bans/" + uID }
	EndpointChannelTopics        = func(cID string) string { return EndpointChannels + cID + "/topics" }
	EndpointChannelTopic         = func(cID, tID string) string { return EndpointChannels + cID + "/topics/" + tID }
	EndpointChannelItems         = func(cID string) string { return EndpointChannels + cID + "/items" }
	EndpointChannelItem          = func(cID, iID string) string { return EndpointChannels + cID + "/items/" + iID }
	EndpointChannelItemComplete  = func(cID, iID string) string { return EndpointChannels + cID + "/items/" + iID + "/complete" }
	EndpointChannelDocs          = func(cID string) string { return EndpointChannels + cID + "/docs" }
	EndpointChannelDoc           = func(cID, dID string) string { return EndpointChannels + cID + "/docs/" + dID }
	EndpointChannelEvents        = func(cID string) string { return EndpointChannels + cID + "/events" }
	EndpointChannelEvent         = func(cID, eID string) string { return EndpointChannels + cID + "/events/" + eID }
	EndpointChannelEventRsvps    = func(cID, eID string) string { return EndpointChannels + cID + "/events/" + eID + "/rsvps" }
	EndpointChannelEventRsvp     = func(cID, eID, uID string) string { return EndpointChannels + cID + "/events/" + eID + "/rsvps/" + uID }
	EndpointChannelReaction      = func(cID, coID, eID string) string {
		return EndpointChannels + cID + "/content/" + coID + "/emotes/" + eID
	}
	EndpointServerXPMember         = func(sID, uID string) string { return EndpointServers + sID + "/members/" + uID + "/xp" }
	EndpointServerXPRoles          = func(sID, rID string) string { return EndpointServers + sID + "/roles/" + rID + "/xp" }
	EndpointServerMemberSocialLink = func(sID, uID, linkType string) string {
		return EndpointServers + sID + "/members/" + uID + "/social-links/" + linkType
	}
	EndpointGroupMember       = func(gID, uID string) string { return EndpointGroups + gID + "/members/" + uID }
	EndpointServerMemberRoles = func(sID, uID string) string { return EndpointServers + sID + "/members/" + uID + "/roles" }
	EndpointServerMemberRole  = func(sID, uID, rID string) string { return EndpointServers + sID + "/members/" + uID + "/roles/" + rID }
	EndpointServerWeebhooks   = func(sID string) string { return EndpointServers + sID + "/webhooks" }
	EndpointServerWeebhook    = func(sID, wID string) string { return EndpointServers + sID + "/webhooks/" + wID }
)
