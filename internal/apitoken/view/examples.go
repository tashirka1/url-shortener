package view

import "fmt"

type apiExample struct {
	Desc string
	Cmd  string
}

func apiExamples(token string) []apiExample {
	return []apiExample{
		{
			Desc: "Создать ссылку",
			Cmd:  fmt.Sprintf("curl -X POST https://YOUR_DOMAIN/api/v1/link \\\n  -H \"Authorization: Bearer %s\" \\\n  -H \"Content-Type: application/json\" \\\n  -d '{\"url\":\"https://example.com\"}'", token),
		},
		{
			Desc: "Список ссылок",
			Cmd:  fmt.Sprintf("curl https://YOUR_DOMAIN/api/v1/link \\\n  -H \"Authorization: Bearer %s\"", token),
		},
		{
			Desc: "Удалить ссылку",
			Cmd:  fmt.Sprintf("curl -X DELETE https://YOUR_DOMAIN/api/v1/link/CODE \\\n  -H \"Authorization: Bearer %s\"", token),
		},
		{
			Desc: "Статистика",
			Cmd:  fmt.Sprintf("curl https://YOUR_DOMAIN/api/v1/link/CODE/stats \\\n  -H \"Authorization: Bearer %s\"", token),
		},
	}
}
