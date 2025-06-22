package main

const (
	// 临时使用绝对路径
	ConfPath = "/Users/sky/GolandProjects/my_redis/data/conf.json"
)

const (
	ColorReset = "\033[0m"

	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
)

const (
	LogInfo  = 1
	LogWarn  = 2
	LogError = 3
)

const (
	TimeLayout = "2006-01-02 15:04:05"
)

const (
	CmdPing = "PING"
	CmdAuth = "AUTH"

	CmdSelect       = "SELECT"
	CmdDBSize       = "DBSIZE"
	CmdBGRewriteAOF = "BGREWRITEAOF"
	CmdSubscribe    = "SUBSCRIBE"
	CmdUnsubscribe  = "UNSUBSCRIBE"
	CmdPublish      = "PUBLISH"

	CmdMulti   = "MULTI"
	CmdDiscard = "DISCARD"
	CmdExec    = "EXEC"
	CmdWatch   = "WATCH"
	CmdUnwatch = "UNWATCH"

	CmdSet    = "SET"
	CmdGet    = "GET"
	CmdIncrBy = "INCRBY"
	CmdSetNX  = "SETNX"
	CmdSetEX  = "SETEX"
	CmdZAdd   = "ZADD"
	CmdZRem   = "ZREM"
	CmdZRange = "ZRANGE"
	CmdZCard  = "ZCARD"
	CmdZScore = "ZSCORE"
	CmdZRank  = "ZRANK"

	CmdExists  = "EXISTS"
	CmdType    = "TYPE"
	CmdTTL     = "TTL"
	CmdDel     = "DEL"
	CmdExpire  = "EXPIRE"
	CmdPersist = "PERSIST"

	CmdAbsExpire = "ABSEXPIRE" // 一般只给系统用 绝对的超时时间，用于 AOF 重放
)

const (
	TypeStr  = 1
	TypeTime = 2
	TypeZSet = 3
)

const (
	FsyncAlways   = "ALWAYS"    // 每次写指令都刷盘
	FsyncEverySec = "EVERY_SEC" // 每秒刷盘一次
	FsyncNo       = "NO"        // 从不刷盘，由操作系统决定
)

const (
	Logo = ColorGreen + "                                                                                \n          ____                                                                  \n        ,'  , `.          ,-.----.                                              \n     ,-+-,.' _ |          \\    /  \\                  ,---,  ,--,                \n  ,-+-. ;   , ||          ;   :    \\               ,---.'|,--.'|                \n ,--.'|'   |  ;|          |   | .\\ :               |   | :|  |,      .--.--.    \n|   |  ,', |  ':     .--, .   : |: |    ,---.      |   | |`--'_     /  /    '   \n|   | /  | |  ||   /_ ./| |   |  \\ :   /     \\   ,--.__| |,' ,'|   |  :  /`./   \n'   | :  | :  |,, ' , ' : |   : .  /  /    /  | /   ,'   |'  | |   |  :  ;_     \n;   . |  ; |--'/___/ \\: | ;   | |  \\ .    ' / |.   '  /  ||  | :    \\  \\    `.  \n|   : |  | ,    .  \\  ' | |   | ;\\  \\'   ;   /|'   ; |:  |'  : |__   `----.   \\ \n|   : '  |/      \\  ;   : :   ' | \\.''   |  / ||   | '/  '|  | '.'| /  /`--'  / \n;   | |`-'        \\  \\  ; :   : :-'  |   :    ||   :    :|;  :    ;'--'.     /  \n|   ;/             :  \\  \\|   |.'     \\   \\  /  \\   \\  /  |  ,   /   `--'---'   \n'---'               \\  ' ;`---'        `----'    `----'    ---`-'               \n                     `--`                                                       " + ColorReset
)
