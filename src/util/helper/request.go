package helper

import "time"

func TimeNow() (*time.Time, error) {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}

	jakartaTime := time.Now().In(location)

	return &jakartaTime, nil
}
