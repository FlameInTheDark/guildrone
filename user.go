package guildrone

import "time"

// User stores the data of a user.
type User struct {
	// The ID of the user
	ID string `json:"id"`

	// The type of the user
	// Can be "bot" or "user"
	Type string `json:"type"`

	// The name of the user
	Name string `json:"name"`

	// The avatar url of the user
	Avatar string `json:"avatar"`

	// The banner url of the user
	Banner string `json:"banner"`

	// The timestamp of when the user was created
	CreatedAt time.Time `json:"createdAt"`
}

// Mention returns a string that can be used to mention the user.
func (u *User) Mention() string {
	return "@" + u.Name + " "
}

// MentionEmbed returns a string that can be used to mention the user in an embed.
func (u *User) MentionEmbed() string {
	return "<@" + u.ID + ">"
}
