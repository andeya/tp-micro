
package types


type User struct {
    userId string
    nickname string
    m1 int
    m2 float
    m3 long
    m4 double
    m5 string
    m6 string
    m7 string
}


type RpcArgs struct {
    a int
    b int
    c string
    userIdList []string
    friendsList [string]*User
}


type RpcReply struct {
    d int
    userList []*User
    userMap [string]*User
}




