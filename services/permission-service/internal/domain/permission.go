package domain

type Action string

const (
    ActionSendMessage   Action = "send_message"
    ActionDeleteMessage Action = "delete_message"
    ActionManageChannel Action = "manage_channel"
    ActionBanMember     Action = "ban_member"
    ActionKickMember    Action = "kick_member"
    ActionManageServer  Action = "manage_server"
    ActionStartSession  Action = "start_session"
    ActionViewChannel   Action = "view_channel"
)

type ContextType string

const (
    ContextGuild   ContextType = "guild"
    ContextChannel ContextType = "channel"
)

type PermissionRequest struct {
    UserID      string
    Role        string
    Action      Action
    ContextType ContextType
    ContextID   string
}

type PermissionResult struct {
    Allowed bool
    Reason  string
}
