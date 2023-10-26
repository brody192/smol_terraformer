package main

import (
	"fmt"
	"main/internal/railway"
	"main/internal/tools"
	"os"
	"sync"
)

func main() {
	var (
		railwayToken  = os.Getenv("RAILWAY_TOKEN")
		projectId     = os.Getenv("RAILWAY_PROJECT_ID")
		EnvironmentId = os.Getenv("RAILWAY_ENVIRONMENT_ID")
	)

	railwayClient := railway.NewAuthedClient(railwayToken)

	serviceCreateResp, err := railway.ServiceCreate(railwayClient, &railway.ServiceCreateInput{
		Name:          "test_" + tools.RandString(5),
		EnvironmentId: EnvironmentId,
		ProjectId:     projectId,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("New service ID: " + serviceCreateResp.ServiceCreate.Id)

	errChan := make(chan error, 1)

	var customDomainCreateResp *railway.CustomDomainCreateResponse

	wg := &sync.WaitGroup{}

	wg.Add(3)

	go func() {
		defer wg.Done()

		resp, err := railway.CustomDomainCreate(railwayClient, &railway.CustomDomainCreateInput{
			Domain:        tools.RandString(10) + ".com",
			EnvironmentId: EnvironmentId,
			ServiceId:     serviceCreateResp.ServiceCreate.Id,
		})

		errChan <- err
		customDomainCreateResp = resp
	}()

	go func() {
		defer wg.Done()

		_, err := railway.VariableCollectionUpsert(railwayClient, &railway.VariableCollectionUpsertInput{
			EnvironmentId: EnvironmentId,
			ServiceId:     serviceCreateResp.ServiceCreate.Id,
			ProjectId:     projectId,
			Replace:       true,
			Variables: map[string]string{
				tools.RandString(5): tools.RandString(5),
				tools.RandString(5): tools.RandString(5),
				tools.RandString(5): tools.RandString(5),
				tools.RandString(5): tools.RandString(5),
			},
		})

		errChan <- err
	}()

	go func() {
		defer wg.Done()

		_, err := railway.ServiceInstanceUpdate(railwayClient, serviceCreateResp.ServiceCreate.Id, EnvironmentId, &railway.ServiceInstanceUpdateInput{
			HealthcheckPath: "/health",
		})

		errChan <- err
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Println("CNAME for custom domain: " + customDomainCreateResp.CustomDomainCreate.Status.DnsRecords[0].Fqdn)

	if _, err := railway.ServiceConnect(railwayClient, serviceCreateResp.ServiceCreate.Id, &railway.ServiceConnectInput{
		Branch: "main",
		Repo:   "brody192/hello-world",
	}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
