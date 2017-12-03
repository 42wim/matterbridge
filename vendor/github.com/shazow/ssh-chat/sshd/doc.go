package sshd

/*

	signer, err := ssh.ParsePrivateKey(privateKey)

	config := MakeNoAuth()
	config.AddHostKey(signer)

	s, err := ListenSSH("0.0.0.0:2022", config)
	if err != nil {
		// Handle opening socket error
	}
	defer s.Close()

	terminals := s.ServeTerminal()

	for term := range terminals {
		go func() {
			defer term.Close()
			term.SetPrompt("...")
			term.AutoCompleteCallback = nil // ...

			for {
				line, err := term.ReadLine()
				if err != nil {
					break
				}
				term.Write(...)
			}

		}()
	}
*/
