package rmq

// func TestRMQ_ConnectWithRetry(t *testing.T) {
// 	rmq := New("amqp://guest:guest@localhost:5672/")
//
// 	err := rmq.Connect()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer rmq.Close()
//
// 	// go func() {
// 	// 	err = rmq.KeepAlive(5, time.Second*5)
// 	// 	if err != nil {
// 	// 		t.Error(err)
// 	// 		return
// 	// 	}
// 	// }()
//
// 	conn, err := rmq.connection(context.Background())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	fmt.Println("connected and sleeping for 10 seconds")
// 	time.Sleep(time.Second * 10)
//
// 	fmt.Println("initiating publisher")
// 	publisher, err := newPublisher(conn, "", "events")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	fmt.Println("publishing")
// 	err = publisher.New().Publish(context.Background(), []byte("Hello"))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Log("published")
//
// 	select {}
// }
