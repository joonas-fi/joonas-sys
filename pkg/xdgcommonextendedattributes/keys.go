// https://www.freedesktop.org/wiki/CommonExtendedAttributes/
package xdgcommonextendedattributes

// These attributes are currently proposed

const (
	Comment              = "user.xdg.comment"                 // A comment specified by the user. This comment could be displayed by file managers
	OriginURL            = "user.xdg.origin.url"              // Set on a file downloaded from a url. Its value should equal the url it was downloaded from.
	OriginEmailSubject   = "user.xdg.origin.email.subject"    // Set on an email attachment when saved to disk. It should get its value from the originating message's "Subject" header
	OriginEmailFrom      = "user.xdg.origin.email.from"       // Set on an email attachment when saved to disk. It should get its value from the originating messsage's "From" header. For example '"John Doe" <jdoe@example.com>', or 'jdoe@example.com'
	OriginEmailMessageID = "user.xdg.origin.email.message-id" // Set on an email attachment when saved to disk. It should get its value from the originating message's "Message-Id" header.
	Language             = "user.xdg.language"                // Language of the intellectual content of the resource. The value should follow the syntax described in RFC 3066 in conjunction with ISO 639 language codes. When a file is downloaded using HTTP, the value of the Content-Language HTTP header can if present be copied into this attribute. See also the Language element in Dublin core.
	Creator              = "user.xdg.creator"                 // Reserved but not yet defined. The string "user" has a different meaning in ROX Contact Manager (creating application) compared with Dublin core (creating person/entity).
	Publisher            = "user.xdg.publisher"               // Name of the creating application. See also the Publisher element in Dublin core.
)

// https://www.freedesktop.org/wiki/CommonExtendedAttributes/#proposedcontrolattributes

const (
	RobotsBackup = "user.xdg.robots.backup"
)
