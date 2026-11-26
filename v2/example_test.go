package interage_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	interage "github.com/WonitTecnologia/interage-sdk-go/v2"
)

func Example() {
	cli, err := interage.NewClient("SEU-TENANT.wonit.cloud", "sk_xxxxxxxxxxxx", nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	fim := time.Now()
	for ligacao, err := range cli.Calls.All(ctx, interage.ListCallsParams{From: fim.AddDate(0, 0, -7), To: fim}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(ligacao.ID, ligacao.StartedAt)
	}
}

func ExampleContactsService_Update() {
	cli, _ := interage.NewClient("SEU-TENANT.wonit.cloud", "sk_xxxxxxxxxxxx", nil)
	_, err := cli.Contacts.Update(context.Background(), "b3f1c2a4-1234-4abc-9def-0123456789ab", interage.ContactUpdate{
		Email:   interage.Set("joao@empresa.com"), // troca
		Company: interage.Null[string](),          // limpa
		Labels:  interage.Set([]int64{}),          // tira todas as etiquetas
	})
	var e *interage.Error
	if errors.As(err, &e) && e.Code == interage.CodeValidationError {
		for _, campo := range e.Errors {
			fmt.Println(campo.Field, campo.Message)
		}
	}
}

func ExampleParseWebhook() {
	http.HandleFunc("/interage", func(w http.ResponseWriter, r *http.Request) {
		lote, err := interage.ParseWebhook(r, os.Getenv("INTERAGE_WEBHOOK_SECRET"), 0)
		if err != nil {
			http.Error(w, "assinatura inválida", http.StatusUnauthorized)
			return
		}
		for _, ev := range lote.Events {
			switch ev.Type {
			case interage.EventMessageCreated:
				fmt.Println(ev.Conversation.Protocol, ev.Message.Type)
			case interage.EventExtensionStatusChanged:
				fmt.Println(ev.Extension.Number, ev.Extension.PreviousStatus, "→", ev.Extension.Status)
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
