package mailservers

import "github.com/status-im/status-go/params"

func DefaultMailserversByFleet(fleet string) []Mailserver {
	var items []Mailserver
	for _, ms := range DefaultMailservers() {
		if ms.Fleet == fleet {
			items = append(items, ms)
		}
	}
	return items
}

func DefaultMailservers() []Mailserver {

	return []Mailserver{
		Mailserver{
			ID:      "mail-01.ac-cn-hongkong-c.eth.prod",
			Address: "enode://606ae04a71e5db868a722c77a21c8244ae38f1bd6e81687cc6cfe88a3063fa1c245692232f64f45bd5408fed5133eab8ed78049332b04f9c110eac7f71c1b429@47.75.247.214:443",
			Fleet:   params.FleetProd,
			Version: 1,
		},
		Mailserver{
			ID:      "mail-01.do-ams3.eth.prod",
			Address: "enode://c42f368a23fa98ee546fd247220759062323249ef657d26d357a777443aec04db1b29a3a22ef3e7c548e18493ddaf51a31b0aed6079bd6ebe5ae838fcfaf3a49@178.128.142.54:443",
			Fleet:   params.FleetProd,
			Version: 1,
		},
		Mailserver{
			ID:      "mail-01.gc-us-central1-a.eth.prod",
			Address: "enode://ee2b53b0ace9692167a410514bca3024695dbf0e1a68e1dff9716da620efb195f04a4b9e873fb9b74ac84de801106c465b8e2b6c4f0d93b8749d1578bfcaf03e@104.197.238.144:443",
			Fleet:   params.FleetProd,
			Version: 1,
		},
		Mailserver{
			ID:      "mail-02.ac-cn-hongkong-c.eth.prod",
			Address: "enode://2c8de3cbb27a3d30cbb5b3e003bc722b126f5aef82e2052aaef032ca94e0c7ad219e533ba88c70585ebd802de206693255335b100307645ab5170e88620d2a81@47.244.221.14:443",
			Fleet:   params.FleetProd,
			Version: 1,
		},
		Mailserver{
			ID:      "mail-02.do-ams3.eth.prod",
			Address: "enode://7aa648d6e855950b2e3d3bf220c496e0cae4adfddef3e1e6062e6b177aec93bc6cdcf1282cb40d1656932ebfdd565729da440368d7c4da7dbd4d004b1ac02bf8@178.128.142.26:443",
			Fleet:   params.FleetProd,
			Version: 1,
		},
		Mailserver{
			ID:      "mail-02.gc-us-central1-a.eth.prod",
			Address: "enode://30211cbd81c25f07b03a0196d56e6ce4604bb13db773ff1c0ea2253547fafd6c06eae6ad3533e2ba39d59564cfbdbb5e2ce7c137a5ebb85e99dcfc7a75f99f55@23.236.58.92:443",
			Fleet:   params.FleetProd,
			Version: 1,
		},
		Mailserver{
			ID:      "mail-03.ac-cn-hongkong-c.eth.prod",
			Address: "enode://e85f1d4209f2f99da801af18db8716e584a28ad0bdc47fbdcd8f26af74dbd97fc279144680553ec7cd9092afe683ddea1e0f9fc571ebcb4b1d857c03a088853d@47.244.129.82:443",
			Fleet:   params.FleetProd,
			Version: 1,
		},
		Mailserver{
			ID:      "mail-03.do-ams3.eth.prod",
			Address: "enode://8a64b3c349a2e0ef4a32ea49609ed6eb3364be1110253c20adc17a3cebbc39a219e5d3e13b151c0eee5d8e0f9a8ba2cd026014e67b41a4ab7d1d5dd67ca27427@178.128.142.94:443",
			Fleet:   params.FleetProd,
			Version: 1,
		},
		Mailserver{
			ID:      "mail-03.gc-us-central1-a.eth.prod",
			Address: "enode://44160e22e8b42bd32a06c1532165fa9e096eebedd7fa6d6e5f8bbef0440bc4a4591fe3651be68193a7ec029021cdb496cfe1d7f9f1dc69eb99226e6f39a7a5d4@35.225.221.245:443",
			Fleet:   params.FleetProd,
			Version: 1,
		},
		Mailserver{
			ID:      "mail-01.ac-cn-hongkong-c.eth.staging",
			Address: "enode://b74859176c9751d314aeeffc26ec9f866a412752e7ddec91b19018a18e7cca8d637cfe2cedcb972f8eb64d816fbd5b4e89c7e8c7fd7df8a1329fa43db80b0bfe@47.52.90.156:443",
			Fleet:   params.FleetStaging,
			Version: 1,
		},
		Mailserver{
			ID:      "mail-01.do-ams3.eth.staging",
			Address: "enode://69f72baa7f1722d111a8c9c68c39a31430e9d567695f6108f31ccb6cd8f0adff4991e7fdca8fa770e75bc8a511a87d24690cbc80e008175f40c157d6f6788d48@206.189.240.16:443",
			Fleet:   params.FleetStaging,
			Version: 1,
		},
		Mailserver{
			ID:      "mail-01.gc-us-central1-a.eth.staging",
			Address: "enode://e4fc10c1f65c8aed83ac26bc1bfb21a45cc1a8550a58077c8d2de2a0e0cd18e40fd40f7e6f7d02dc6cd06982b014ce88d6e468725ffe2c138e958788d0002a7f@35.239.193.41:443",
			Fleet:   params.FleetStaging,
			Version: 1,
		},
		Mailserver{
			ID:      "mail-01.ac-cn-hongkong-c.eth.test",
			Address: "enode://619dbb5dda12e85bf0eb5db40fb3de625609043242737c0e975f7dfd659d85dc6d9a84f9461a728c5ab68c072fed38ca6a53917ca24b8e93cc27bdef3a1e79ac@47.52.188.196:443",
			Fleet:   params.FleetTest,
			Version: 1,
		},
		Mailserver{
			ID:      "mail-01.do-ams3.eth.test",
			Address: "enode://e4865fe6c2a9c1a563a6447990d8e9ce672644ae3e08277ce38ec1f1b690eef6320c07a5d60c3b629f5d4494f93d6b86a745a0bf64ab295bbf6579017adc6ed8@206.189.243.161:443",
			Fleet:   params.FleetTest,
			Version: 1,
		},
		Mailserver{
			ID:      "mail-01.gc-us-central1-a.eth.test",
			Address: "enode://707e57453acd3e488c44b9d0e17975371e2f8fb67525eae5baca9b9c8e06c86cde7c794a6c2e36203bf9f56cae8b0e50f3b33c4c2b694a7baeea1754464ce4e3@35.192.229.172:443",
			Fleet:   params.FleetTest,
			Version: 1,
		},
		Mailserver{
			ID:      "node-01.ac-cn-hongkong-c.wakuv2.prod",
			Address: "/ip4/8.210.222.231/tcp/30303/p2p/16Uiu2HAm4v86W3bmT1BiH6oSPzcsSr24iDQpSN5Qa992BCjjwgrD",
			Fleet:   params.FleetWakuV2Prod,
			Version: 2,
		},
		Mailserver{
			ID:      "node-01.do-ams3.wakuv2.prod",
			Address: "/ip4/188.166.135.145/tcp/30303/p2p/16Uiu2HAmL5okWopX7NqZWBUKVqW8iUxCEmd5GMHLVPwCgzYzQv3e",
			Fleet:   params.FleetWakuV2Prod,
			Version: 2,
		},
		Mailserver{
			ID:      "node-01.gc-us-central1-a.wakuv2.prod",
			Address: "/ip4/34.121.100.108/tcp/30303/p2p/16Uiu2HAmVkKntsECaYfefR1V2yCR79CegLATuTPE6B9TxgxBiiiA",
			Fleet:   params.FleetWakuV2Prod,
			Version: 2,
		},
		Mailserver{
			ID:      "node-01.ac-cn-hongkong-c.wakuv2.test",
			Address: "/ip4/47.242.210.73/tcp/30303/p2p/16Uiu2HAkvWiyFsgRhuJEb9JfjYxEkoHLgnUQmr1N5mKWnYjxYRVm",
			Fleet:   params.FleetWakuV2Test,
			Version: 2,
		},
		Mailserver{
			ID:      "node-01.do-ams3.wakuv2.test",
			Address: "/ip4/134.209.139.210/tcp/30303/p2p/16Uiu2HAmPLe7Mzm8TsYUubgCAW1aJoeFScxrLj8ppHFivPo97bUZ",
			Fleet:   params.FleetWakuV2Test,
			Version: 2,
		},
		Mailserver{
			ID:      "node-01.gc-us-central1-a.wakuv2.test",
			Address: "/ip4/104.154.239.128/tcp/30303/p2p/16Uiu2HAmJb2e28qLXxT5kZxVUUoJt72EMzNGXB47Rxx5hw3q4YjS",
			Fleet:   params.FleetWakuV2Test,
			Version: 2,
		},
		Mailserver{
			ID:      "node-01.ac-cn-hongkong-c.status.prod",
			Address: "/dns4/node-01.ac-cn-hongkong-c.status.prod.statusim.net/tcp/30303/p2p/16Uiu2HAkvEZgh3KLwhLwXg95e5ojM8XykJ4Kxi2T7hk22rnA7pJC",
			Fleet:   params.FleetStatusProd,
			Version: 2,
		},
		Mailserver{
			ID:      "node-01.do-ams3.status.prod",
			Address: "/dns4/node-01.do-ams3.status.prod.statusim.net/tcp/30303/p2p/16Uiu2HAm6HZZr7aToTvEBPpiys4UxajCTU97zj5v7RNR2gbniy1D",
			Fleet:   params.FleetStatusProd,
			Version: 2,
		},
		Mailserver{
			ID:      "node-01.gc-us-central1-a.status.prod",
			Address: "/dns4/node-01.gc-us-central1-a.status.prod.statusim.net/tcp/30303/p2p/16Uiu2HAkwBp8T6G77kQXSNMnxgaMky1JeyML5yqoTHRM8dbeCBNb",
			Fleet:   params.FleetStatusProd,
			Version: 2,
		},
		Mailserver{
			ID:      "node-02.ac-cn-hongkong-c.status.prod",
			Address: "/dns4/node-02.ac-cn-hongkong-c.status.prod.statusim.net/tcp/30303/p2p/16Uiu2HAmFy8BrJhCEmCYrUfBdSNkrPw6VHExtv4rRp1DSBnCPgx8",
			Fleet:   params.FleetStatusProd,
			Version: 2,
		},
		Mailserver{
			ID:      "node-02.do-ams3.status.prod",
			Address: "/dns4/node-02.do-ams3.status.prod.statusim.net/tcp/30303/p2p/16Uiu2HAmSve7tR5YZugpskMv2dmJAsMUKmfWYEKRXNUxRaTCnsXV",
			Fleet:   params.FleetStatusProd,
			Version: 2,
		},
		Mailserver{
			ID:      "node-02.gc-us-central1-a.status.prod",
			Address: "/dns4/node-02.gc-us-central1-a.status.prod.statusim.net/tcp/30303/p2p/16Uiu2HAmDQugwDHM3YeUp86iGjrUvbdw3JPRgikC7YoGBsT2ymMg",
			Fleet:   params.FleetStatusProd,
			Version: 2,
		},
		Mailserver{
			ID:      "node-01.ac-cn-hongkong-c.status.test",
			Address: "/dns4/node-01.ac-cn-hongkong-c.status.test.statusim.net/tcp/30303/p2p/16Uiu2HAm2BjXxCp1sYFJQKpLLbPbwd5juxbsYofu3TsS3auvT9Yi",
			Fleet:   params.FleetStatusTest,
			Version: 2,
		},
		Mailserver{
			ID:      "node-01.do-ams3.status.test",
			Address: "/dns4/node-01.do-ams3.status.test.statusim.net/tcp/30303/p2p/16Uiu2HAkukebeXjTQ9QDBeNDWuGfbaSg79wkkhK4vPocLgR6QFDf",
			Fleet:   params.FleetStatusTest,
			Version: 2,
		},
		Mailserver{
			ID:      "node-01.gc-us-central1-a.status.test",
			Address: "/dns4/node-01.gc-us-central1-a.status.test.statusim.net/tcp/30303/p2p/16Uiu2HAmGDX3iAFox93PupVYaHa88kULGqMpJ7AEHGwj3jbMtt76",
			Fleet:   params.FleetStatusTest,
			Version: 2,
		},
		Mailserver{
			ID:      "store-01.do-ams3.shards.test",
			Address: "/dns4/store-01.do-ams3.shards.test.statusim.net/tcp/30303/p2p/16Uiu2HAmAUdrQ3uwzuE4Gy4D56hX6uLKEeerJAnhKEHZ3DxF1EfT",
			Fleet:   params.FleetShardsTest,
			Version: 2,
		},
		Mailserver{
			ID:      "store-02.do-ams3.shards.test",
			Address: "/dns4/store-02.do-ams3.shards.test.statusim.net/tcp/30303/p2p/16Uiu2HAm9aDJPkhGxc2SFcEACTFdZ91Q5TJjp76qZEhq9iF59x7R",
			Fleet:   params.FleetShardsTest,
			Version: 2,
		},
		Mailserver{
			ID:      "store-01.gc-us-central1-a.shards.test",
			Address: "/dns4/store-01.gc-us-central1-a.shards.test.statusim.net/tcp/30303/p2p/16Uiu2HAmMELCo218hncCtTvC2Dwbej3rbyHQcR8erXNnKGei7WPZ",
			Fleet:   params.FleetShardsTest,
			Version: 2,
		},
		Mailserver{
			ID:      "store-02.gc-us-central1-a.shards.test",
			Address: "/dns4/store-02.gc-us-central1-a.shards.test.statusim.net/tcp/30303/p2p/16Uiu2HAmJnVR7ZzFaYvciPVafUXuYGLHPzSUigqAmeNw9nJUVGeM",
			Fleet:   params.FleetShardsTest,
			Version: 2,
		},
		Mailserver{
			ID:      "store-01.ac-cn-hongkong-c.shards.test",
			Address: "/dns4/store-01.ac-cn-hongkong-c.shards.test.statusim.net/tcp/30303/p2p/16Uiu2HAm2M7xs7cLPc3jamawkEqbr7cUJX11uvY7LxQ6WFUdUKUT",
			Fleet:   params.FleetShardsTest,
			Version: 2,
		},
		Mailserver{
			ID:      "store-02.ac-cn-hongkong-c.shards.test",
			Address: "/dns4/store-02.ac-cn-hongkong-c.shards.test.statusim.net/tcp/30303/p2p/16Uiu2HAm9CQhsuwPR54q27kNj9iaQVfyRzTGKrhFmr94oD8ujU6P",
			Fleet:   params.FleetShardsTest,
			Version: 2,
		},
	}
}
