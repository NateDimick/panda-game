let userInfo: UserInfoType = $state({Name: null, SessionID: null, PlayerID: null})

export let UserInfo = {
    update: (newState: UserInfoType) => { userInfo = newState},
    get name() {return userInfo.Name},
    get id() {return userInfo.PlayerID}
}

export type UserInfoType = {
    Name: string | null
    SessionID: string | null
    PlayerID: string | null
}