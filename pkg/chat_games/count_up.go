package chat_games

import (
    "log"
    "strconv"

    database "github.com/replit/database-go"
)

const CountUpName = "CountUp"

type CountUp struct {
    Score uint64
}

func NewCountUp() (CountUp) {
    return CountUp {
        Score: 0,
    }
}

func (c CountUp) ReadScore() (uint64, error) {
    dbValue, err := database.Get(CountUpName)
    if err != nil {
        return 0, err
    }

    convertedValue, err := strconv.ParseUint(dbValue, 10, 64)
    if err != nil {
        log.Fatal(err)
        return 0, err
    }

    return convertedValue, nil
}

func (c CountUp) WriteScore() {
    database.Set(CountUpName, strconv.FormatUint(c.Score, 10))
}

func (c CountUp) StoreHighScore() {
    lastHighScore, err := c.ReadScore()
    if err != nil {
        log.Fatal(err)
        return
    }

    if c.Score > lastHighScore {
        c.WriteScore()
    }
}
