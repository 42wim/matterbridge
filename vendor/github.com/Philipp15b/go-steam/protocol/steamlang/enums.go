// Generated code
// DO NOT EDIT

package steamlang

import (
	"fmt"
	"sort"
	"strings"
)

type EMsg int32

const (
	EMsg_Invalid                                                  EMsg = 0
	EMsg_Multi                                                    EMsg = 1
	EMsg_ProtobufWrapped                                          EMsg = 2
	EMsg_BaseGeneral                                              EMsg = 100
	EMsg_GenericReply                                             EMsg = 100
	EMsg_DestJobFailed                                            EMsg = 113
	EMsg_Alert                                                    EMsg = 115
	EMsg_SCIDRequest                                              EMsg = 120
	EMsg_SCIDResponse                                             EMsg = 121
	EMsg_JobHeartbeat                                             EMsg = 123
	EMsg_HubConnect                                               EMsg = 124
	EMsg_Subscribe                                                EMsg = 126
	EMsg_RouteMessage                                             EMsg = 127
	EMsg_AMCreateAccountResponse                                  EMsg = 129 // Deprecated
	EMsg_WGRequest                                                EMsg = 130
	EMsg_WGResponse                                               EMsg = 131
	EMsg_KeepAlive                                                EMsg = 132
	EMsg_WebAPIJobRequest                                         EMsg = 133
	EMsg_WebAPIJobResponse                                        EMsg = 134
	EMsg_ClientSessionStart                                       EMsg = 135
	EMsg_ClientSessionEnd                                         EMsg = 136
	EMsg_ClientSessionUpdate                                      EMsg = 137
	EMsg_Ping                                                     EMsg = 139
	EMsg_PingResponse                                             EMsg = 140
	EMsg_Stats                                                    EMsg = 141
	EMsg_RequestFullStatsBlock                                    EMsg = 142
	EMsg_LoadDBOCacheItem                                         EMsg = 143
	EMsg_LoadDBOCacheItemResponse                                 EMsg = 144
	EMsg_InvalidateDBOCacheItems                                  EMsg = 145
	EMsg_ServiceMethod                                            EMsg = 146
	EMsg_ServiceMethodResponse                                    EMsg = 147
	EMsg_ClientPackageVersions                                    EMsg = 148
	EMsg_TimestampRequest                                         EMsg = 149
	EMsg_TimestampResponse                                        EMsg = 150
	EMsg_ServiceMethodCallFromClient                              EMsg = 151
	EMsg_ServiceMethodSendToClient                                EMsg = 152
	EMsg_BaseShell                                                EMsg = 200
	EMsg_AssignSysID                                              EMsg = 200
	EMsg_Exit                                                     EMsg = 201
	EMsg_DirRequest                                               EMsg = 202
	EMsg_DirResponse                                              EMsg = 203
	EMsg_ZipRequest                                               EMsg = 204
	EMsg_ZipResponse                                              EMsg = 205
	EMsg_UpdateRecordResponse                                     EMsg = 215
	EMsg_UpdateCreditCardRequest                                  EMsg = 221
	EMsg_UpdateUserBanResponse                                    EMsg = 225
	EMsg_PrepareToExit                                            EMsg = 226
	EMsg_ContentDescriptionUpdate                                 EMsg = 227
	EMsg_TestResetServer                                          EMsg = 228
	EMsg_UniverseChanged                                          EMsg = 229
	EMsg_ShellConfigInfoUpdate                                    EMsg = 230
	EMsg_RequestWindowsEventLogEntries                            EMsg = 233
	EMsg_ProvideWindowsEventLogEntries                            EMsg = 234
	EMsg_ShellSearchLogs                                          EMsg = 235
	EMsg_ShellSearchLogsResponse                                  EMsg = 236
	EMsg_ShellCheckWindowsUpdates                                 EMsg = 237
	EMsg_ShellCheckWindowsUpdatesResponse                         EMsg = 238
	EMsg_TestFlushDelayedSQL                                      EMsg = 240
	EMsg_TestFlushDelayedSQLResponse                              EMsg = 241
	EMsg_EnsureExecuteScheduledTask_TEST                          EMsg = 242
	EMsg_EnsureExecuteScheduledTaskResponse_TEST                  EMsg = 243
	EMsg_UpdateScheduledTaskEnableState_TEST                      EMsg = 244
	EMsg_UpdateScheduledTaskEnableStateResponse_TEST              EMsg = 245
	EMsg_ContentDescriptionDeltaUpdate                            EMsg = 246
	EMsg_BaseGM                                                   EMsg = 300
	EMsg_Heartbeat                                                EMsg = 300
	EMsg_ShellFailed                                              EMsg = 301
	EMsg_ExitShells                                               EMsg = 307
	EMsg_ExitShell                                                EMsg = 308
	EMsg_GracefulExitShell                                        EMsg = 309
	EMsg_LicenseProcessingComplete                                EMsg = 316
	EMsg_SetTestFlag                                              EMsg = 317
	EMsg_QueuedEmailsComplete                                     EMsg = 318
	EMsg_GMReportPHPError                                         EMsg = 319
	EMsg_GMDRMSync                                                EMsg = 320
	EMsg_PhysicalBoxInventory                                     EMsg = 321
	EMsg_UpdateConfigFile                                         EMsg = 322
	EMsg_TestInitDB                                               EMsg = 323
	EMsg_GMWriteConfigToSQL                                       EMsg = 324
	EMsg_GMLoadActivationCodes                                    EMsg = 325
	EMsg_GMQueueForFBS                                            EMsg = 326
	EMsg_GMSchemaConversionResults                                EMsg = 327
	EMsg_GMWriteShellFailureToSQL                                 EMsg = 329
	EMsg_GMWriteStatsToSOS                                        EMsg = 330
	EMsg_GMGetServiceMethodRouting                                EMsg = 331
	EMsg_GMGetServiceMethodRoutingResponse                        EMsg = 332
	EMsg_GMConvertUserWallets                                     EMsg = 333
	EMsg_GMTestNextBuildSchemaConversion                          EMsg = 334
	EMsg_GMTestNextBuildSchemaConversionResponse                  EMsg = 335
	EMsg_ExpectShellRestart                                       EMsg = 336
	EMsg_HotFixProgress                                           EMsg = 337
	EMsg_BaseAIS                                                  EMsg = 400
	EMsg_AISRequestContentDescription                             EMsg = 402
	EMsg_AISUpdateAppInfo                                         EMsg = 403
	EMsg_AISUpdatePackageCosts                                    EMsg = 404 // Deprecated
	EMsg_AISGetPackageChangeNumber                                EMsg = 405
	EMsg_AISGetPackageChangeNumberResponse                        EMsg = 406
	EMsg_AISUpdatePackageCostsResponse                            EMsg = 408 // Deprecated
	EMsg_AISCreateMarketingMessage                                EMsg = 409 // Deprecated
	EMsg_AISCreateMarketingMessageResponse                        EMsg = 410 // Deprecated
	EMsg_AISGetMarketingMessage                                   EMsg = 411 // Deprecated
	EMsg_AISGetMarketingMessageResponse                           EMsg = 412 // Deprecated
	EMsg_AISUpdateMarketingMessage                                EMsg = 413 // Deprecated
	EMsg_AISUpdateMarketingMessageResponse                        EMsg = 414 // Deprecated
	EMsg_AISRequestMarketingMessageUpdate                         EMsg = 415 // Deprecated
	EMsg_AISDeleteMarketingMessage                                EMsg = 416 // Deprecated
	EMsg_AIGetAppGCFlags                                          EMsg = 423
	EMsg_AIGetAppGCFlagsResponse                                  EMsg = 424
	EMsg_AIGetAppList                                             EMsg = 425
	EMsg_AIGetAppListResponse                                     EMsg = 426
	EMsg_AISGetCouponDefinition                                   EMsg = 429
	EMsg_AISGetCouponDefinitionResponse                           EMsg = 430
	EMsg_AISUpdateSlaveContentDescription                         EMsg = 431
	EMsg_AISUpdateSlaveContentDescriptionResponse                 EMsg = 432
	EMsg_AISTestEnableGC                                          EMsg = 433
	EMsg_BaseAM                                                   EMsg = 500
	EMsg_AMUpdateUserBanRequest                                   EMsg = 504
	EMsg_AMAddLicense                                             EMsg = 505
	EMsg_AMSendSystemIMToUser                                     EMsg = 508
	EMsg_AMExtendLicense                                          EMsg = 509
	EMsg_AMAddMinutesToLicense                                    EMsg = 510
	EMsg_AMCancelLicense                                          EMsg = 511
	EMsg_AMInitPurchase                                           EMsg = 512
	EMsg_AMPurchaseResponse                                       EMsg = 513
	EMsg_AMGetFinalPrice                                          EMsg = 514
	EMsg_AMGetFinalPriceResponse                                  EMsg = 515
	EMsg_AMGetLegacyGameKey                                       EMsg = 516
	EMsg_AMGetLegacyGameKeyResponse                               EMsg = 517
	EMsg_AMFindHungTransactions                                   EMsg = 518
	EMsg_AMSetAccountTrustedRequest                               EMsg = 519
	EMsg_AMCompletePurchase                                       EMsg = 521 // Deprecated
	EMsg_AMCancelPurchase                                         EMsg = 522
	EMsg_AMNewChallenge                                           EMsg = 523
	EMsg_AMLoadOEMTickets                                         EMsg = 524
	EMsg_AMFixPendingPurchase                                     EMsg = 525
	EMsg_AMFixPendingPurchaseResponse                             EMsg = 526
	EMsg_AMIsUserBanned                                           EMsg = 527
	EMsg_AMRegisterKey                                            EMsg = 528
	EMsg_AMLoadActivationCodes                                    EMsg = 529
	EMsg_AMLoadActivationCodesResponse                            EMsg = 530
	EMsg_AMLookupKeyResponse                                      EMsg = 531
	EMsg_AMLookupKey                                              EMsg = 532
	EMsg_AMChatCleanup                                            EMsg = 533
	EMsg_AMClanCleanup                                            EMsg = 534
	EMsg_AMFixPendingRefund                                       EMsg = 535
	EMsg_AMReverseChargeback                                      EMsg = 536
	EMsg_AMReverseChargebackResponse                              EMsg = 537
	EMsg_AMClanCleanupList                                        EMsg = 538
	EMsg_AMGetLicenses                                            EMsg = 539
	EMsg_AMGetLicensesResponse                                    EMsg = 540
	EMsg_AMSendCartRepurchase                                     EMsg = 541
	EMsg_AMSendCartRepurchaseResponse                             EMsg = 542
	EMsg_AllowUserToPlayQuery                                     EMsg = 550
	EMsg_AllowUserToPlayResponse                                  EMsg = 551
	EMsg_AMVerfiyUser                                             EMsg = 552
	EMsg_AMClientNotPlaying                                       EMsg = 553
	EMsg_ClientRequestFriendship                                  EMsg = 554 // Deprecated: Renamed to AMClientRequestFriendship
	EMsg_AMClientRequestFriendship                                EMsg = 554
	EMsg_AMRelayPublishStatus                                     EMsg = 555
	EMsg_AMInitPurchaseResponse                                   EMsg = 560
	EMsg_AMRevokePurchaseResponse                                 EMsg = 561
	EMsg_AMRefreshGuestPasses                                     EMsg = 563
	EMsg_AMInviteUserToClan                                       EMsg = 564 // Deprecated
	EMsg_AMAcknowledgeClanInvite                                  EMsg = 565 // Deprecated
	EMsg_AMGrantGuestPasses                                       EMsg = 566
	EMsg_AMClanDataUpdated                                        EMsg = 567
	EMsg_AMReloadAccount                                          EMsg = 568
	EMsg_AMClientChatMsgRelay                                     EMsg = 569
	EMsg_AMChatMulti                                              EMsg = 570
	EMsg_AMClientChatInviteRelay                                  EMsg = 571
	EMsg_AMChatInvite                                             EMsg = 572
	EMsg_AMClientJoinChatRelay                                    EMsg = 573
	EMsg_AMClientChatMemberInfoRelay                              EMsg = 574
	EMsg_AMPublishChatMemberInfo                                  EMsg = 575
	EMsg_AMClientAcceptFriendInvite                               EMsg = 576
	EMsg_AMChatEnter                                              EMsg = 577
	EMsg_AMClientPublishRemovalFromSource                         EMsg = 578
	EMsg_AMChatActionResult                                       EMsg = 579
	EMsg_AMFindAccounts                                           EMsg = 580
	EMsg_AMFindAccountsResponse                                   EMsg = 581
	EMsg_AMRequestAccountData                                     EMsg = 582
	EMsg_AMRequestAccountDataResponse                             EMsg = 583
	EMsg_AMSetAccountFlags                                        EMsg = 584
	EMsg_AMCreateClan                                             EMsg = 586
	EMsg_AMCreateClanResponse                                     EMsg = 587
	EMsg_AMGetClanDetails                                         EMsg = 588
	EMsg_AMGetClanDetailsResponse                                 EMsg = 589
	EMsg_AMSetPersonaName                                         EMsg = 590
	EMsg_AMSetAvatar                                              EMsg = 591
	EMsg_AMAuthenticateUser                                       EMsg = 592
	EMsg_AMAuthenticateUserResponse                               EMsg = 593
	EMsg_AMP2PIntroducerMessage                                   EMsg = 596
	EMsg_ClientChatAction                                         EMsg = 597
	EMsg_AMClientChatActionRelay                                  EMsg = 598
	EMsg_BaseVS                                                   EMsg = 600
	EMsg_ReqChallenge                                             EMsg = 600
	EMsg_VACResponse                                              EMsg = 601
	EMsg_ReqChallengeTest                                         EMsg = 602
	EMsg_VSMarkCheat                                              EMsg = 604
	EMsg_VSAddCheat                                               EMsg = 605
	EMsg_VSPurgeCodeModDB                                         EMsg = 606
	EMsg_VSGetChallengeResults                                    EMsg = 607
	EMsg_VSChallengeResultText                                    EMsg = 608
	EMsg_VSReportLingerer                                         EMsg = 609
	EMsg_VSRequestManagedChallenge                                EMsg = 610
	EMsg_VSLoadDBFinished                                         EMsg = 611
	EMsg_BaseDRMS                                                 EMsg = 625
	EMsg_DRMBuildBlobRequest                                      EMsg = 628
	EMsg_DRMBuildBlobResponse                                     EMsg = 629
	EMsg_DRMResolveGuidRequest                                    EMsg = 630
	EMsg_DRMResolveGuidResponse                                   EMsg = 631
	EMsg_DRMVariabilityReport                                     EMsg = 633
	EMsg_DRMVariabilityReportResponse                             EMsg = 634
	EMsg_DRMStabilityReport                                       EMsg = 635
	EMsg_DRMStabilityReportResponse                               EMsg = 636
	EMsg_DRMDetailsReportRequest                                  EMsg = 637
	EMsg_DRMDetailsReportResponse                                 EMsg = 638
	EMsg_DRMProcessFile                                           EMsg = 639
	EMsg_DRMAdminUpdate                                           EMsg = 640
	EMsg_DRMAdminUpdateResponse                                   EMsg = 641
	EMsg_DRMSync                                                  EMsg = 642
	EMsg_DRMSyncResponse                                          EMsg = 643
	EMsg_DRMProcessFileResponse                                   EMsg = 644
	EMsg_DRMEmptyGuidCache                                        EMsg = 645
	EMsg_DRMEmptyGuidCacheResponse                                EMsg = 646
	EMsg_BaseCS                                                   EMsg = 650
	EMsg_BaseClient                                               EMsg = 700
	EMsg_ClientHeartBeat                                          EMsg = 703
	EMsg_ClientVACResponse                                        EMsg = 704
	EMsg_ClientLogOff                                             EMsg = 706
	EMsg_ClientNoUDPConnectivity                                  EMsg = 707
	EMsg_ClientInformOfCreateAccount                              EMsg = 708 // Deprecated
	EMsg_ClientConnectionStats                                    EMsg = 710
	EMsg_ClientPingResponse                                       EMsg = 712
	EMsg_ClientRemoveFriend                                       EMsg = 714
	EMsg_ClientGamesPlayedNoDataBlob                              EMsg = 715
	EMsg_ClientChangeStatus                                       EMsg = 716
	EMsg_ClientVacStatusResponse                                  EMsg = 717
	EMsg_ClientFriendMsg                                          EMsg = 718
	EMsg_ClientSystemIM                                           EMsg = 726
	EMsg_ClientSystemIMAck                                        EMsg = 727
	EMsg_ClientGetLicenses                                        EMsg = 728
	EMsg_ClientGetLegacyGameKey                                   EMsg = 730
	EMsg_ClientAckVACBan2                                         EMsg = 732
	EMsg_ClientGetPurchaseReceipts                                EMsg = 736
	EMsg_ClientAckGuestPass                                       EMsg = 740
	EMsg_ClientRedeemGuestPass                                    EMsg = 741
	EMsg_ClientGamesPlayed                                        EMsg = 742
	EMsg_ClientRegisterKey                                        EMsg = 743
	EMsg_ClientInviteUserToClan                                   EMsg = 744
	EMsg_ClientAcknowledgeClanInvite                              EMsg = 745
	EMsg_ClientPurchaseWithMachineID                              EMsg = 746
	EMsg_ClientAppUsageEvent                                      EMsg = 747
	EMsg_ClientLogOnResponse                                      EMsg = 751
	EMsg_ClientSetHeartbeatRate                                   EMsg = 755
	EMsg_ClientLoggedOff                                          EMsg = 757
	EMsg_GSApprove                                                EMsg = 758
	EMsg_GSDeny                                                   EMsg = 759
	EMsg_GSKick                                                   EMsg = 760
	EMsg_ClientCreateAcctResponse                                 EMsg = 761
	EMsg_ClientPurchaseResponse                                   EMsg = 763
	EMsg_ClientPing                                               EMsg = 764
	EMsg_ClientNOP                                                EMsg = 765
	EMsg_ClientPersonaState                                       EMsg = 766
	EMsg_ClientFriendsList                                        EMsg = 767
	EMsg_ClientAccountInfo                                        EMsg = 768
	EMsg_ClientNewsUpdate                                         EMsg = 771
	EMsg_ClientGameConnectDeny                                    EMsg = 773
	EMsg_GSStatusReply                                            EMsg = 774
	EMsg_ClientGameConnectTokens                                  EMsg = 779
	EMsg_ClientLicenseList                                        EMsg = 780
	EMsg_ClientVACBanStatus                                       EMsg = 782
	EMsg_ClientCMList                                             EMsg = 783
	EMsg_ClientEncryptPct                                         EMsg = 784
	EMsg_ClientGetLegacyGameKeyResponse                           EMsg = 785
	EMsg_ClientAddFriend                                          EMsg = 791
	EMsg_ClientAddFriendResponse                                  EMsg = 792
	EMsg_ClientAckGuestPassResponse                               EMsg = 796
	EMsg_ClientRedeemGuestPassResponse                            EMsg = 797
	EMsg_ClientUpdateGuestPassesList                              EMsg = 798
	EMsg_ClientChatMsg                                            EMsg = 799
	EMsg_ClientChatInvite                                         EMsg = 800
	EMsg_ClientJoinChat                                           EMsg = 801
	EMsg_ClientChatMemberInfo                                     EMsg = 802
	EMsg_ClientPasswordChangeResponse                             EMsg = 805
	EMsg_ClientChatEnter                                          EMsg = 807
	EMsg_ClientFriendRemovedFromSource                            EMsg = 808
	EMsg_ClientCreateChat                                         EMsg = 809
	EMsg_ClientCreateChatResponse                                 EMsg = 810
	EMsg_ClientP2PIntroducerMessage                               EMsg = 813
	EMsg_ClientChatActionResult                                   EMsg = 814
	EMsg_ClientRequestFriendData                                  EMsg = 815
	EMsg_ClientGetUserStats                                       EMsg = 818
	EMsg_ClientGetUserStatsResponse                               EMsg = 819
	EMsg_ClientStoreUserStats                                     EMsg = 820
	EMsg_ClientStoreUserStatsResponse                             EMsg = 821
	EMsg_ClientClanState                                          EMsg = 822
	EMsg_ClientServiceModule                                      EMsg = 830
	EMsg_ClientServiceCall                                        EMsg = 831
	EMsg_ClientServiceCallResponse                                EMsg = 832
	EMsg_ClientNatTraversalStatEvent                              EMsg = 839
	EMsg_ClientSteamUsageEvent                                    EMsg = 842
	EMsg_ClientCheckPassword                                      EMsg = 845
	EMsg_ClientResetPassword                                      EMsg = 846
	EMsg_ClientCheckPasswordResponse                              EMsg = 848
	EMsg_ClientResetPasswordResponse                              EMsg = 849
	EMsg_ClientSessionToken                                       EMsg = 850
	EMsg_ClientDRMProblemReport                                   EMsg = 851
	EMsg_ClientSetIgnoreFriend                                    EMsg = 855
	EMsg_ClientSetIgnoreFriendResponse                            EMsg = 856
	EMsg_ClientGetAppOwnershipTicket                              EMsg = 857
	EMsg_ClientGetAppOwnershipTicketResponse                      EMsg = 858
	EMsg_ClientGetLobbyListResponse                               EMsg = 860
	EMsg_ClientServerList                                         EMsg = 880
	EMsg_ClientEmailChangeResponse                                EMsg = 891
	EMsg_ClientSecretQAChangeResponse                             EMsg = 892
	EMsg_ClientDRMBlobRequest                                     EMsg = 896
	EMsg_ClientDRMBlobResponse                                    EMsg = 897
	EMsg_BaseGameServer                                           EMsg = 900
	EMsg_GSDisconnectNotice                                       EMsg = 901
	EMsg_GSStatus                                                 EMsg = 903
	EMsg_GSUserPlaying                                            EMsg = 905
	EMsg_GSStatus2                                                EMsg = 906
	EMsg_GSStatusUpdate_Unused                                    EMsg = 907
	EMsg_GSServerType                                             EMsg = 908
	EMsg_GSPlayerList                                             EMsg = 909
	EMsg_GSGetUserAchievementStatus                               EMsg = 910
	EMsg_GSGetUserAchievementStatusResponse                       EMsg = 911
	EMsg_GSGetPlayStats                                           EMsg = 918
	EMsg_GSGetPlayStatsResponse                                   EMsg = 919
	EMsg_GSGetUserGroupStatus                                     EMsg = 920
	EMsg_AMGetUserGroupStatus                                     EMsg = 921
	EMsg_AMGetUserGroupStatusResponse                             EMsg = 922
	EMsg_GSGetUserGroupStatusResponse                             EMsg = 923
	EMsg_GSGetReputation                                          EMsg = 936
	EMsg_GSGetReputationResponse                                  EMsg = 937
	EMsg_GSAssociateWithClan                                      EMsg = 938
	EMsg_GSAssociateWithClanResponse                              EMsg = 939
	EMsg_GSComputeNewPlayerCompatibility                          EMsg = 940
	EMsg_GSComputeNewPlayerCompatibilityResponse                  EMsg = 941
	EMsg_BaseAdmin                                                EMsg = 1000
	EMsg_AdminCmd                                                 EMsg = 1000
	EMsg_AdminCmdResponse                                         EMsg = 1004
	EMsg_AdminLogListenRequest                                    EMsg = 1005
	EMsg_AdminLogEvent                                            EMsg = 1006
	EMsg_UniverseData                                             EMsg = 1010
	EMsg_AdminSpew                                                EMsg = 1019
	EMsg_AdminConsoleTitle                                        EMsg = 1020
	EMsg_AdminGCSpew                                              EMsg = 1023
	EMsg_AdminGCCommand                                           EMsg = 1024
	EMsg_AdminGCGetCommandList                                    EMsg = 1025
	EMsg_AdminGCGetCommandListResponse                            EMsg = 1026
	EMsg_FBSConnectionData                                        EMsg = 1027
	EMsg_AdminMsgSpew                                             EMsg = 1028
	EMsg_BaseFBS                                                  EMsg = 1100
	EMsg_FBSReqVersion                                            EMsg = 1100
	EMsg_FBSVersionInfo                                           EMsg = 1101
	EMsg_FBSForceRefresh                                          EMsg = 1102
	EMsg_FBSForceBounce                                           EMsg = 1103
	EMsg_FBSDeployPackage                                         EMsg = 1104
	EMsg_FBSDeployResponse                                        EMsg = 1105
	EMsg_FBSUpdateBootstrapper                                    EMsg = 1106
	EMsg_FBSSetState                                              EMsg = 1107
	EMsg_FBSApplyOSUpdates                                        EMsg = 1108
	EMsg_FBSRunCMDScript                                          EMsg = 1109
	EMsg_FBSRebootBox                                             EMsg = 1110
	EMsg_FBSSetBigBrotherMode                                     EMsg = 1111
	EMsg_FBSMinidumpServer                                        EMsg = 1112
	EMsg_FBSDeployHotFixPackage                                   EMsg = 1114
	EMsg_FBSDeployHotFixResponse                                  EMsg = 1115
	EMsg_FBSDownloadHotFix                                        EMsg = 1116
	EMsg_FBSDownloadHotFixResponse                                EMsg = 1117
	EMsg_FBSUpdateTargetConfigFile                                EMsg = 1118
	EMsg_FBSApplyAccountCred                                      EMsg = 1119
	EMsg_FBSApplyAccountCredResponse                              EMsg = 1120
	EMsg_FBSSetShellCount                                         EMsg = 1121
	EMsg_FBSTerminateShell                                        EMsg = 1122
	EMsg_FBSQueryGMForRequest                                     EMsg = 1123
	EMsg_FBSQueryGMResponse                                       EMsg = 1124
	EMsg_FBSTerminateZombies                                      EMsg = 1125
	EMsg_FBSInfoFromBootstrapper                                  EMsg = 1126
	EMsg_FBSRebootBoxResponse                                     EMsg = 1127
	EMsg_FBSBootstrapperPackageRequest                            EMsg = 1128
	EMsg_FBSBootstrapperPackageResponse                           EMsg = 1129
	EMsg_FBSBootstrapperGetPackageChunk                           EMsg = 1130
	EMsg_FBSBootstrapperGetPackageChunkResponse                   EMsg = 1131
	EMsg_FBSBootstrapperPackageTransferProgress                   EMsg = 1132
	EMsg_FBSRestartBootstrapper                                   EMsg = 1133
	EMsg_FBSPauseFrozenDumps                                      EMsg = 1134
	EMsg_BaseFileXfer                                             EMsg = 1200
	EMsg_FileXferRequest                                          EMsg = 1200
	EMsg_FileXferResponse                                         EMsg = 1201
	EMsg_FileXferData                                             EMsg = 1202
	EMsg_FileXferEnd                                              EMsg = 1203
	EMsg_FileXferDataAck                                          EMsg = 1204
	EMsg_BaseChannelAuth                                          EMsg = 1300
	EMsg_ChannelAuthChallenge                                     EMsg = 1300
	EMsg_ChannelAuthResponse                                      EMsg = 1301
	EMsg_ChannelAuthResult                                        EMsg = 1302
	EMsg_ChannelEncryptRequest                                    EMsg = 1303
	EMsg_ChannelEncryptResponse                                   EMsg = 1304
	EMsg_ChannelEncryptResult                                     EMsg = 1305
	EMsg_BaseBS                                                   EMsg = 1400
	EMsg_BSPurchaseStart                                          EMsg = 1401
	EMsg_BSPurchaseResponse                                       EMsg = 1402
	EMsg_BSAuthenticateCCTrans                                    EMsg = 1403
	EMsg_BSAuthenticateCCTransResponse                            EMsg = 1404
	EMsg_BSSettleComplete                                         EMsg = 1406
	EMsg_BSInitPayPalTxn                                          EMsg = 1408
	EMsg_BSInitPayPalTxnResponse                                  EMsg = 1409
	EMsg_BSGetPayPalUserInfo                                      EMsg = 1410
	EMsg_BSGetPayPalUserInfoResponse                              EMsg = 1411
	EMsg_BSPaymentInstrBan                                        EMsg = 1417
	EMsg_BSPaymentInstrBanResponse                                EMsg = 1418
	EMsg_BSInitGCBankXferTxn                                      EMsg = 1421
	EMsg_BSInitGCBankXferTxnResponse                              EMsg = 1422
	EMsg_BSCommitGCTxn                                            EMsg = 1425
	EMsg_BSQueryTransactionStatus                                 EMsg = 1426
	EMsg_BSQueryTransactionStatusResponse                         EMsg = 1427
	EMsg_BSQueryPaymentInstUsage                                  EMsg = 1431
	EMsg_BSQueryPaymentInstResponse                               EMsg = 1432
	EMsg_BSQueryTxnExtendedInfo                                   EMsg = 1433
	EMsg_BSQueryTxnExtendedInfoResponse                           EMsg = 1434
	EMsg_BSUpdateConversionRates                                  EMsg = 1435
	EMsg_BSPurchaseRunFraudChecks                                 EMsg = 1437
	EMsg_BSPurchaseRunFraudChecksResponse                         EMsg = 1438
	EMsg_BSQueryBankInformation                                   EMsg = 1440
	EMsg_BSQueryBankInformationResponse                           EMsg = 1441
	EMsg_BSValidateXsollaSignature                                EMsg = 1445
	EMsg_BSValidateXsollaSignatureResponse                        EMsg = 1446
	EMsg_BSQiwiWalletInvoice                                      EMsg = 1448
	EMsg_BSQiwiWalletInvoiceResponse                              EMsg = 1449
	EMsg_BSUpdateInventoryFromProPack                             EMsg = 1450
	EMsg_BSUpdateInventoryFromProPackResponse                     EMsg = 1451
	EMsg_BSSendShippingRequest                                    EMsg = 1452
	EMsg_BSSendShippingRequestResponse                            EMsg = 1453
	EMsg_BSGetProPackOrderStatus                                  EMsg = 1454
	EMsg_BSGetProPackOrderStatusResponse                          EMsg = 1455
	EMsg_BSCheckJobRunning                                        EMsg = 1456
	EMsg_BSCheckJobRunningResponse                                EMsg = 1457
	EMsg_BSResetPackagePurchaseRateLimit                          EMsg = 1458
	EMsg_BSResetPackagePurchaseRateLimitResponse                  EMsg = 1459
	EMsg_BSUpdatePaymentData                                      EMsg = 1460
	EMsg_BSUpdatePaymentDataResponse                              EMsg = 1461
	EMsg_BSGetBillingAddress                                      EMsg = 1462
	EMsg_BSGetBillingAddressResponse                              EMsg = 1463
	EMsg_BSGetCreditCardInfo                                      EMsg = 1464
	EMsg_BSGetCreditCardInfoResponse                              EMsg = 1465
	EMsg_BSRemoveExpiredPaymentData                               EMsg = 1468
	EMsg_BSRemoveExpiredPaymentDataResponse                       EMsg = 1469
	EMsg_BSConvertToCurrentKeys                                   EMsg = 1470
	EMsg_BSConvertToCurrentKeysResponse                           EMsg = 1471
	EMsg_BSInitPurchase                                           EMsg = 1472
	EMsg_BSInitPurchaseResponse                                   EMsg = 1473
	EMsg_BSCompletePurchase                                       EMsg = 1474
	EMsg_BSCompletePurchaseResponse                               EMsg = 1475
	EMsg_BSPruneCardUsageStats                                    EMsg = 1476
	EMsg_BSPruneCardUsageStatsResponse                            EMsg = 1477
	EMsg_BSStoreBankInformation                                   EMsg = 1478
	EMsg_BSStoreBankInformationResponse                           EMsg = 1479
	EMsg_BSVerifyPOSAKey                                          EMsg = 1480
	EMsg_BSVerifyPOSAKeyResponse                                  EMsg = 1481
	EMsg_BSReverseRedeemPOSAKey                                   EMsg = 1482
	EMsg_BSReverseRedeemPOSAKeyResponse                           EMsg = 1483
	EMsg_BSQueryFindCreditCard                                    EMsg = 1484
	EMsg_BSQueryFindCreditCardResponse                            EMsg = 1485
	EMsg_BSStatusInquiryPOSAKey                                   EMsg = 1486
	EMsg_BSStatusInquiryPOSAKeyResponse                           EMsg = 1487
	EMsg_BSValidateMoPaySignature                                 EMsg = 1488 // Deprecated
	EMsg_BSValidateMoPaySignatureResponse                         EMsg = 1489 // Deprecated
	EMsg_BSMoPayConfirmProductDelivery                            EMsg = 1490 // Deprecated
	EMsg_BSMoPayConfirmProductDeliveryResponse                    EMsg = 1491 // Deprecated
	EMsg_BSGenerateMoPayMD5                                       EMsg = 1492 // Deprecated
	EMsg_BSGenerateMoPayMD5Response                               EMsg = 1493 // Deprecated
	EMsg_BSBoaCompraConfirmProductDelivery                        EMsg = 1494
	EMsg_BSBoaCompraConfirmProductDeliveryResponse                EMsg = 1495
	EMsg_BSGenerateBoaCompraMD5                                   EMsg = 1496
	EMsg_BSGenerateBoaCompraMD5Response                           EMsg = 1497
	EMsg_BSCommitWPTxn                                            EMsg = 1498
	EMsg_BSCommitAdyenTxn                                         EMsg = 1499
	EMsg_BaseATS                                                  EMsg = 1500
	EMsg_ATSStartStressTest                                       EMsg = 1501
	EMsg_ATSStopStressTest                                        EMsg = 1502
	EMsg_ATSRunFailServerTest                                     EMsg = 1503
	EMsg_ATSUFSPerfTestTask                                       EMsg = 1504
	EMsg_ATSUFSPerfTestResponse                                   EMsg = 1505
	EMsg_ATSCycleTCM                                              EMsg = 1506
	EMsg_ATSInitDRMSStressTest                                    EMsg = 1507
	EMsg_ATSCallTest                                              EMsg = 1508
	EMsg_ATSCallTestReply                                         EMsg = 1509
	EMsg_ATSStartExternalStress                                   EMsg = 1510
	EMsg_ATSExternalStressJobStart                                EMsg = 1511
	EMsg_ATSExternalStressJobQueued                               EMsg = 1512
	EMsg_ATSExternalStressJobRunning                              EMsg = 1513
	EMsg_ATSExternalStressJobStopped                              EMsg = 1514
	EMsg_ATSExternalStressJobStopAll                              EMsg = 1515
	EMsg_ATSExternalStressActionResult                            EMsg = 1516
	EMsg_ATSStarted                                               EMsg = 1517
	EMsg_ATSCSPerfTestTask                                        EMsg = 1518
	EMsg_ATSCSPerfTestResponse                                    EMsg = 1519
	EMsg_BaseDP                                                   EMsg = 1600
	EMsg_DPSetPublishingState                                     EMsg = 1601
	EMsg_DPUniquePlayersStat                                      EMsg = 1603
	EMsg_DPStreamingUniquePlayersStat                             EMsg = 1604
	EMsg_DPVacInfractionStats                                     EMsg = 1605 // Deprecated
	EMsg_DPVacBanStats                                            EMsg = 1606 // Deprecated
	EMsg_DPBlockingStats                                          EMsg = 1607
	EMsg_DPNatTraversalStats                                      EMsg = 1608
	EMsg_DPVacCertBanStats                                        EMsg = 1610 // Deprecated
	EMsg_DPVacCafeBanStats                                        EMsg = 1611 // Deprecated
	EMsg_DPCloudStats                                             EMsg = 1612
	EMsg_DPAchievementStats                                       EMsg = 1613
	EMsg_DPAccountCreationStats                                   EMsg = 1614 // Deprecated
	EMsg_DPGetPlayerCount                                         EMsg = 1615
	EMsg_DPGetPlayerCountResponse                                 EMsg = 1616
	EMsg_DPGameServersPlayersStats                                EMsg = 1617
	EMsg_DPFacebookStatistics                                     EMsg = 1619 // Deprecated
	EMsg_ClientDPCheckSpecialSurvey                               EMsg = 1620
	EMsg_ClientDPCheckSpecialSurveyResponse                       EMsg = 1621
	EMsg_ClientDPSendSpecialSurveyResponse                        EMsg = 1622
	EMsg_ClientDPSendSpecialSurveyResponseReply                   EMsg = 1623
	EMsg_DPStoreSaleStatistics                                    EMsg = 1624
	EMsg_ClientDPUpdateAppJobReport                               EMsg = 1625
	EMsg_DPUpdateContentEvent                                     EMsg = 1626
	EMsg_ClientDPUnsignedInstallScript                            EMsg = 1627
	EMsg_DPPartnerMicroTxns                                       EMsg = 1628
	EMsg_DPPartnerMicroTxnsResponse                               EMsg = 1629
	EMsg_ClientDPContentStatsReport                               EMsg = 1630
	EMsg_DPVRUniquePlayersStat                                    EMsg = 1631
	EMsg_BaseCM                                                   EMsg = 1700
	EMsg_CMSetAllowState                                          EMsg = 1701
	EMsg_CMSpewAllowState                                         EMsg = 1702
	EMsg_CMSessionRejected                                        EMsg = 1703
	EMsg_CMSetSecrets                                             EMsg = 1704
	EMsg_CMGetSecrets                                             EMsg = 1705
	EMsg_BaseGC                                                   EMsg = 2200
	EMsg_GCCmdRevive                                              EMsg = 2203
	EMsg_GCCmdDown                                                EMsg = 2206
	EMsg_GCCmdDeploy                                              EMsg = 2207
	EMsg_GCCmdDeployResponse                                      EMsg = 2208
	EMsg_GCCmdSwitch                                              EMsg = 2209
	EMsg_AMRefreshSessions                                        EMsg = 2210
	EMsg_GCAchievementAwarded                                     EMsg = 2212
	EMsg_GCSystemMessage                                          EMsg = 2213
	EMsg_GCCmdStatus                                              EMsg = 2216
	EMsg_GCInterAppMessage                                        EMsg = 2219
	EMsg_GCGetEmailTemplate                                       EMsg = 2220
	EMsg_GCGetEmailTemplateResponse                               EMsg = 2221
	EMsg_GCHRelay                                                 EMsg = 2222
	EMsg_GCHRelayToClient                                         EMsg = 2223
	EMsg_GCHUpdateSession                                         EMsg = 2224
	EMsg_GCHRequestUpdateSession                                  EMsg = 2225
	EMsg_GCHRequestStatus                                         EMsg = 2226
	EMsg_GCHRequestStatusResponse                                 EMsg = 2227
	EMsg_GCHAccountVacStatusChange                                EMsg = 2228
	EMsg_GCHSpawnGC                                               EMsg = 2229
	EMsg_GCHSpawnGCResponse                                       EMsg = 2230
	EMsg_GCHKillGC                                                EMsg = 2231
	EMsg_GCHKillGCResponse                                        EMsg = 2232
	EMsg_GCHAccountTradeBanStatusChange                           EMsg = 2233
	EMsg_GCHAccountLockStatusChange                               EMsg = 2234
	EMsg_GCHVacVerificationChange                                 EMsg = 2235
	EMsg_GCHAccountPhoneNumberChange                              EMsg = 2236
	EMsg_GCHAccountTwoFactorChange                                EMsg = 2237
	EMsg_GCHInviteUserToLobby                                     EMsg = 2238
	EMsg_BaseP2P                                                  EMsg = 2500
	EMsg_P2PIntroducerMessage                                     EMsg = 2502
	EMsg_BaseSM                                                   EMsg = 2900
	EMsg_SMExpensiveReport                                        EMsg = 2902
	EMsg_SMHourlyReport                                           EMsg = 2903
	EMsg_SMFishingReport                                          EMsg = 2904 // Deprecated
	EMsg_SMPartitionRenames                                       EMsg = 2905
	EMsg_SMMonitorSpace                                           EMsg = 2906
	EMsg_SMTestNextBuildSchemaConversion                          EMsg = 2907
	EMsg_SMTestNextBuildSchemaConversionResponse                  EMsg = 2908
	EMsg_BaseTest                                                 EMsg = 3000
	EMsg_FailServer                                               EMsg = 3000
	EMsg_JobHeartbeatTest                                         EMsg = 3001
	EMsg_JobHeartbeatTestResponse                                 EMsg = 3002
	EMsg_BaseFTSRange                                             EMsg = 3100
	EMsg_BaseCCSRange                                             EMsg = 3150
	EMsg_CCSDeleteAllCommentsByAuthor                             EMsg = 3161
	EMsg_CCSDeleteAllCommentsByAuthorResponse                     EMsg = 3162
	EMsg_BaseLBSRange                                             EMsg = 3200
	EMsg_LBSSetScore                                              EMsg = 3201
	EMsg_LBSSetScoreResponse                                      EMsg = 3202
	EMsg_LBSFindOrCreateLB                                        EMsg = 3203
	EMsg_LBSFindOrCreateLBResponse                                EMsg = 3204
	EMsg_LBSGetLBEntries                                          EMsg = 3205
	EMsg_LBSGetLBEntriesResponse                                  EMsg = 3206
	EMsg_LBSGetLBList                                             EMsg = 3207
	EMsg_LBSGetLBListResponse                                     EMsg = 3208
	EMsg_LBSSetLBDetails                                          EMsg = 3209
	EMsg_LBSDeleteLB                                              EMsg = 3210
	EMsg_LBSDeleteLBEntry                                         EMsg = 3211
	EMsg_LBSResetLB                                               EMsg = 3212
	EMsg_LBSResetLBResponse                                       EMsg = 3213
	EMsg_LBSDeleteLBResponse                                      EMsg = 3214
	EMsg_BaseOGS                                                  EMsg = 3400
	EMsg_OGSBeginSession                                          EMsg = 3401
	EMsg_OGSBeginSessionResponse                                  EMsg = 3402
	EMsg_OGSEndSession                                            EMsg = 3403
	EMsg_OGSEndSessionResponse                                    EMsg = 3404
	EMsg_OGSWriteAppSessionRow                                    EMsg = 3406
	EMsg_BaseBRP                                                  EMsg = 3600
	EMsg_BRPStartShippingJobs                                     EMsg = 3601 // Deprecated
	EMsg_BRPProcessUSBankReports                                  EMsg = 3602 // Deprecated
	EMsg_BRPProcessGCReports                                      EMsg = 3603 // Deprecated
	EMsg_BRPProcessPPReports                                      EMsg = 3604 // Deprecated
	EMsg_BRPCommitGC                                              EMsg = 3607 // Deprecated
	EMsg_BRPCommitGCResponse                                      EMsg = 3608 // Deprecated
	EMsg_BRPFindHungTransactions                                  EMsg = 3609 // Deprecated
	EMsg_BRPCheckFinanceCloseOutDate                              EMsg = 3610 // Deprecated
	EMsg_BRPProcessLicenses                                       EMsg = 3611 // Deprecated
	EMsg_BRPProcessLicensesResponse                               EMsg = 3612 // Deprecated
	EMsg_BRPRemoveExpiredPaymentData                              EMsg = 3613 // Deprecated
	EMsg_BRPRemoveExpiredPaymentDataResponse                      EMsg = 3614 // Deprecated
	EMsg_BRPConvertToCurrentKeys                                  EMsg = 3615 // Deprecated
	EMsg_BRPConvertToCurrentKeysResponse                          EMsg = 3616 // Deprecated
	EMsg_BRPPruneCardUsageStats                                   EMsg = 3617 // Deprecated
	EMsg_BRPPruneCardUsageStatsResponse                           EMsg = 3618 // Deprecated
	EMsg_BRPCheckActivationCodes                                  EMsg = 3619 // Deprecated
	EMsg_BRPCheckActivationCodesResponse                          EMsg = 3620 // Deprecated
	EMsg_BRPCommitWP                                              EMsg = 3621 // Deprecated
	EMsg_BRPCommitWPResponse                                      EMsg = 3622 // Deprecated
	EMsg_BRPProcessWPReports                                      EMsg = 3623 // Deprecated
	EMsg_BRPProcessPaymentRules                                   EMsg = 3624 // Deprecated
	EMsg_BRPProcessPartnerPayments                                EMsg = 3625 // Deprecated
	EMsg_BRPCheckSettlementReports                                EMsg = 3626 // Deprecated
	EMsg_BRPPostTaxToAvalara                                      EMsg = 3628 // Deprecated
	EMsg_BRPPostTransactionTax                                    EMsg = 3629
	EMsg_BRPPostTransactionTaxResponse                            EMsg = 3630
	EMsg_BRPProcessIMReports                                      EMsg = 3631 // Deprecated
	EMsg_BaseAMRange2                                             EMsg = 4000
	EMsg_AMCreateChat                                             EMsg = 4001
	EMsg_AMCreateChatResponse                                     EMsg = 4002
	EMsg_AMSetProfileURL                                          EMsg = 4005
	EMsg_AMGetAccountEmailAddress                                 EMsg = 4006
	EMsg_AMGetAccountEmailAddressResponse                         EMsg = 4007
	EMsg_AMRequestClanData                                        EMsg = 4008
	EMsg_AMRouteToClients                                         EMsg = 4009
	EMsg_AMLeaveClan                                              EMsg = 4010
	EMsg_AMClanPermissions                                        EMsg = 4011
	EMsg_AMClanPermissionsResponse                                EMsg = 4012
	EMsg_AMCreateClanEvent                                        EMsg = 4013 // Deprecated: renamed to AMCreateClanEventDummyForRateLimiting
	EMsg_AMCreateClanEventDummyForRateLimiting                    EMsg = 4013
	EMsg_AMCreateClanEventResponse                                EMsg = 4014 // Deprecated
	EMsg_AMUpdateClanEvent                                        EMsg = 4015 // Deprecated: renamed to AMUpdateClanEventDummyForRateLimiting
	EMsg_AMUpdateClanEventDummyForRateLimiting                    EMsg = 4015
	EMsg_AMUpdateClanEventResponse                                EMsg = 4016 // Deprecated
	EMsg_AMGetClanEvents                                          EMsg = 4017 // Deprecated
	EMsg_AMGetClanEventsResponse                                  EMsg = 4018 // Deprecated
	EMsg_AMDeleteClanEvent                                        EMsg = 4019 // Deprecated
	EMsg_AMDeleteClanEventResponse                                EMsg = 4020 // Deprecated
	EMsg_AMSetClanPermissionSettings                              EMsg = 4021
	EMsg_AMSetClanPermissionSettingsResponse                      EMsg = 4022
	EMsg_AMGetClanPermissionSettings                              EMsg = 4023
	EMsg_AMGetClanPermissionSettingsResponse                      EMsg = 4024
	EMsg_AMPublishChatRoomInfo                                    EMsg = 4025
	EMsg_ClientChatRoomInfo                                       EMsg = 4026
	EMsg_AMGetClanHistory                                         EMsg = 4039
	EMsg_AMGetClanHistoryResponse                                 EMsg = 4040
	EMsg_AMGetClanPermissionBits                                  EMsg = 4041
	EMsg_AMGetClanPermissionBitsResponse                          EMsg = 4042
	EMsg_AMSetClanPermissionBits                                  EMsg = 4043
	EMsg_AMSetClanPermissionBitsResponse                          EMsg = 4044
	EMsg_AMSessionInfoRequest                                     EMsg = 4045
	EMsg_AMSessionInfoResponse                                    EMsg = 4046
	EMsg_AMValidateWGToken                                        EMsg = 4047
	EMsg_AMGetSingleClanEvent                                     EMsg = 4048 // Deprecated
	EMsg_AMGetSingleClanEventResponse                             EMsg = 4049 // Deprecated
	EMsg_AMGetClanRank                                            EMsg = 4050
	EMsg_AMGetClanRankResponse                                    EMsg = 4051
	EMsg_AMSetClanRank                                            EMsg = 4052
	EMsg_AMSetClanRankResponse                                    EMsg = 4053
	EMsg_AMGetClanPOTW                                            EMsg = 4054
	EMsg_AMGetClanPOTWResponse                                    EMsg = 4055
	EMsg_AMSetClanPOTW                                            EMsg = 4056
	EMsg_AMSetClanPOTWResponse                                    EMsg = 4057
	EMsg_AMDumpUser                                               EMsg = 4059
	EMsg_AMKickUserFromClan                                       EMsg = 4060
	EMsg_AMAddFounderToClan                                       EMsg = 4061
	EMsg_AMValidateWGTokenResponse                                EMsg = 4062
	EMsg_AMSetCommunityState                                      EMsg = 4063 // Deprecated
	EMsg_AMSetAccountDetails                                      EMsg = 4064
	EMsg_AMGetChatBanList                                         EMsg = 4065
	EMsg_AMGetChatBanListResponse                                 EMsg = 4066
	EMsg_AMUnBanFromChat                                          EMsg = 4067
	EMsg_AMSetClanDetails                                         EMsg = 4068
	EMsg_AMGetAccountLinks                                        EMsg = 4069
	EMsg_AMGetAccountLinksResponse                                EMsg = 4070
	EMsg_AMSetAccountLinks                                        EMsg = 4071
	EMsg_AMSetAccountLinksResponse                                EMsg = 4072
	EMsg_AMGetUserGameStats                                       EMsg = 4073 // Deprecated: renamed to UGSGetUserGameStats
	EMsg_UGSGetUserGameStats                                      EMsg = 4073
	EMsg_AMGetUserGameStatsResponse                               EMsg = 4074 // Deprecated: renamed to UGSGetUserGameStatsResponse
	EMsg_UGSGetUserGameStatsResponse                              EMsg = 4074
	EMsg_AMCheckClanMembership                                    EMsg = 4075
	EMsg_AMGetClanMembers                                         EMsg = 4076
	EMsg_AMGetClanMembersResponse                                 EMsg = 4077
	EMsg_AMJoinPublicClan                                         EMsg = 4078 // Deprecated
	EMsg_AMNotifyChatOfClanChange                                 EMsg = 4079
	EMsg_AMResubmitPurchase                                       EMsg = 4080
	EMsg_AMAddFriend                                              EMsg = 4081
	EMsg_AMAddFriendResponse                                      EMsg = 4082
	EMsg_AMRemoveFriend                                           EMsg = 4083
	EMsg_AMDumpClan                                               EMsg = 4084
	EMsg_AMChangeClanOwner                                        EMsg = 4085
	EMsg_AMCancelEasyCollect                                      EMsg = 4086
	EMsg_AMCancelEasyCollectResponse                              EMsg = 4087
	EMsg_AMClansInCommon                                          EMsg = 4090
	EMsg_AMClansInCommonResponse                                  EMsg = 4091
	EMsg_AMIsValidAccountID                                       EMsg = 4092
	EMsg_AMConvertClan                                            EMsg = 4093 // Deprecated
	EMsg_AMWipeFriendsList                                        EMsg = 4095
	EMsg_AMSetIgnored                                             EMsg = 4096
	EMsg_AMClansInCommonCountResponse                             EMsg = 4097
	EMsg_AMFriendsList                                            EMsg = 4098
	EMsg_AMFriendsListResponse                                    EMsg = 4099
	EMsg_AMFriendsInCommon                                        EMsg = 4100
	EMsg_AMFriendsInCommonResponse                                EMsg = 4101
	EMsg_AMFriendsInCommonCountResponse                           EMsg = 4102
	EMsg_AMClansInCommonCount                                     EMsg = 4103
	EMsg_AMChallengeVerdict                                       EMsg = 4104
	EMsg_AMChallengeNotification                                  EMsg = 4105
	EMsg_AMFindGSByIP                                             EMsg = 4106
	EMsg_AMFoundGSByIP                                            EMsg = 4107
	EMsg_AMGiftRevoked                                            EMsg = 4108
	EMsg_AMCreateAccountRecord                                    EMsg = 4109 // Deprecated
	EMsg_AMUserClanList                                           EMsg = 4110
	EMsg_AMUserClanListResponse                                   EMsg = 4111
	EMsg_AMGetAccountDetails2                                     EMsg = 4112
	EMsg_AMGetAccountDetailsResponse2                             EMsg = 4113
	EMsg_AMSetCommunityProfileSettings                            EMsg = 4114
	EMsg_AMSetCommunityProfileSettingsResponse                    EMsg = 4115
	EMsg_AMGetCommunityPrivacyState                               EMsg = 4116
	EMsg_AMGetCommunityPrivacyStateResponse                       EMsg = 4117
	EMsg_AMCheckClanInviteRateLimiting                            EMsg = 4118
	EMsg_AMGetUserAchievementStatus                               EMsg = 4119 // Deprecated: renamed to UGSGetUserAchievementStatus
	EMsg_UGSGetUserAchievementStatus                              EMsg = 4119
	EMsg_AMGetIgnored                                             EMsg = 4120
	EMsg_AMGetIgnoredResponse                                     EMsg = 4121
	EMsg_AMSetIgnoredResponse                                     EMsg = 4122
	EMsg_AMSetFriendRelationshipNone                              EMsg = 4123
	EMsg_AMGetFriendRelationship                                  EMsg = 4124
	EMsg_AMGetFriendRelationshipResponse                          EMsg = 4125
	EMsg_AMServiceModulesCache                                    EMsg = 4126
	EMsg_AMServiceModulesCall                                     EMsg = 4127
	EMsg_AMServiceModulesCallResponse                             EMsg = 4128
	EMsg_AMGetCaptchaDataForIP                                    EMsg = 4129 // Deprecated
	EMsg_AMGetCaptchaDataForIPResponse                            EMsg = 4130 // Deprecated
	EMsg_AMValidateCaptchaDataForIP                               EMsg = 4131 // Deprecated
	EMsg_AMValidateCaptchaDataForIPResponse                       EMsg = 4132 // Deprecated
	EMsg_AMTrackFailedAuthByIP                                    EMsg = 4133 // Deprecated
	EMsg_AMGetCaptchaDataByGID                                    EMsg = 4134 // Deprecated
	EMsg_AMGetCaptchaDataByGIDResponse                            EMsg = 4135 // Deprecated
	EMsg_CommunityAddFriendNews                                   EMsg = 4140
	EMsg_AMFindClanUser                                           EMsg = 4143
	EMsg_AMFindClanUserResponse                                   EMsg = 4144
	EMsg_AMBanFromChat                                            EMsg = 4145
	EMsg_AMGetUserNewsSubscriptions                               EMsg = 4147
	EMsg_AMGetUserNewsSubscriptionsResponse                       EMsg = 4148
	EMsg_AMSetUserNewsSubscriptions                               EMsg = 4149
	EMsg_AMSendQueuedEmails                                       EMsg = 4152
	EMsg_AMSetLicenseFlags                                        EMsg = 4153
	EMsg_CommunityDeleteUserNews                                  EMsg = 4155
	EMsg_AMAllowUserFilesRequest                                  EMsg = 4156
	EMsg_AMAllowUserFilesResponse                                 EMsg = 4157
	EMsg_AMGetAccountStatus                                       EMsg = 4158
	EMsg_AMGetAccountStatusResponse                               EMsg = 4159
	EMsg_AMEditBanReason                                          EMsg = 4160
	EMsg_AMCheckClanMembershipResponse                            EMsg = 4161
	EMsg_AMProbeClanMembershipList                                EMsg = 4162
	EMsg_AMProbeClanMembershipListResponse                        EMsg = 4163
	EMsg_UGSGetUserAchievementStatusResponse                      EMsg = 4164
	EMsg_AMGetFriendsLobbies                                      EMsg = 4165
	EMsg_AMGetFriendsLobbiesResponse                              EMsg = 4166
	EMsg_AMGetUserFriendNewsResponse                              EMsg = 4172
	EMsg_CommunityGetUserFriendNews                               EMsg = 4173
	EMsg_AMGetUserClansNewsResponse                               EMsg = 4174
	EMsg_AMGetUserClansNews                                       EMsg = 4175
	EMsg_AMGetPreviousCBAccount                                   EMsg = 4184
	EMsg_AMGetPreviousCBAccountResponse                           EMsg = 4185
	EMsg_AMGetUserLicenseHistory                                  EMsg = 4190
	EMsg_AMGetUserLicenseHistoryResponse                          EMsg = 4191
	EMsg_AMSupportChangePassword                                  EMsg = 4194
	EMsg_AMSupportChangeEmail                                     EMsg = 4195
	EMsg_AMResetUserVerificationGSByIP                            EMsg = 4197
	EMsg_AMUpdateGSPlayStats                                      EMsg = 4198
	EMsg_AMSupportEnableOrDisable                                 EMsg = 4199
	EMsg_AMGetPurchaseStatus                                      EMsg = 4206
	EMsg_AMSupportIsAccountEnabled                                EMsg = 4209
	EMsg_AMSupportIsAccountEnabledResponse                        EMsg = 4210
	EMsg_AMGetUserStats                                           EMsg = 4211 // Deprecated: renamed to UGSGetUserStats
	EMsg_UGSGetUserStats                                          EMsg = 4211
	EMsg_AMSupportKickSession                                     EMsg = 4212
	EMsg_AMGSSearch                                               EMsg = 4213
	EMsg_MarketingMessageUpdate                                   EMsg = 4216
	EMsg_AMRouteFriendMsg                                         EMsg = 4219 // Deprecated: renamed to ChatServerRouteFriendMsg
	EMsg_ChatServerRouteFriendMsg                                 EMsg = 4219
	EMsg_AMTicketAuthRequestOrResponse                            EMsg = 4220
	EMsg_AMVerifyDepotManagementRights                            EMsg = 4222
	EMsg_AMVerifyDepotManagementRightsResponse                    EMsg = 4223
	EMsg_AMAddFreeLicense                                         EMsg = 4224
	EMsg_AMValidateEmailLink                                      EMsg = 4231
	EMsg_AMValidateEmailLinkResponse                              EMsg = 4232
	EMsg_AMStoreUserStats                                         EMsg = 4236 // Deprecated: renamed to UGSStoreUserStats
	EMsg_UGSStoreUserStats                                        EMsg = 4236
	EMsg_AMDeleteStoredCard                                       EMsg = 4241
	EMsg_AMRevokeLegacyGameKeys                                   EMsg = 4242
	EMsg_AMGetWalletDetails                                       EMsg = 4244
	EMsg_AMGetWalletDetailsResponse                               EMsg = 4245
	EMsg_AMDeleteStoredPaymentInfo                                EMsg = 4246
	EMsg_AMGetStoredPaymentSummary                                EMsg = 4247
	EMsg_AMGetStoredPaymentSummaryResponse                        EMsg = 4248
	EMsg_AMGetWalletConversionRate                                EMsg = 4249
	EMsg_AMGetWalletConversionRateResponse                        EMsg = 4250
	EMsg_AMConvertWallet                                          EMsg = 4251
	EMsg_AMConvertWalletResponse                                  EMsg = 4252
	EMsg_AMSetPreApproval                                         EMsg = 4255
	EMsg_AMSetPreApprovalResponse                                 EMsg = 4256
	EMsg_AMCreateRefund                                           EMsg = 4258
	EMsg_AMCreateRefundResponse                                   EMsg = 4259 // Deprecated
	EMsg_AMCreateChargeback                                       EMsg = 4260
	EMsg_AMCreateChargebackResponse                               EMsg = 4261 // Deprecated
	EMsg_AMCreateDispute                                          EMsg = 4262
	EMsg_AMCreateDisputeResponse                                  EMsg = 4263 // Deprecated
	EMsg_AMClearDispute                                           EMsg = 4264
	EMsg_AMClearDisputeResponse                                   EMsg = 4265 // Deprecated: renamed to AMCreateFinancialAdjustment
	EMsg_AMCreateFinancialAdjustment                              EMsg = 4265
	EMsg_AMPlayerNicknameList                                     EMsg = 4266
	EMsg_AMPlayerNicknameListResponse                             EMsg = 4267
	EMsg_AMSetDRMTestConfig                                       EMsg = 4268
	EMsg_AMGetUserCurrentGameInfo                                 EMsg = 4269
	EMsg_AMGetUserCurrentGameInfoResponse                         EMsg = 4270
	EMsg_AMGetGSPlayerList                                        EMsg = 4271
	EMsg_AMGetGSPlayerListResponse                                EMsg = 4272
	EMsg_AMGetGameMembers                                         EMsg = 4276
	EMsg_AMGetGameMembersResponse                                 EMsg = 4277
	EMsg_AMGetSteamIDForMicroTxn                                  EMsg = 4278
	EMsg_AMGetSteamIDForMicroTxnResponse                          EMsg = 4279
	EMsg_AMAddPublisherUser                                       EMsg = 4280 // Deprecated: renamed to AMSetPartnerMember
	EMsg_AMSetPartnerMember                                       EMsg = 4280
	EMsg_AMRemovePublisherUser                                    EMsg = 4281
	EMsg_AMGetUserLicenseList                                     EMsg = 4282
	EMsg_AMGetUserLicenseListResponse                             EMsg = 4283
	EMsg_AMReloadGameGroupPolicy                                  EMsg = 4284
	EMsg_AMAddFreeLicenseResponse                                 EMsg = 4285
	EMsg_AMVACStatusUpdate                                        EMsg = 4286
	EMsg_AMGetAccountDetails                                      EMsg = 4287
	EMsg_AMGetAccountDetailsResponse                              EMsg = 4288
	EMsg_AMGetPlayerLinkDetails                                   EMsg = 4289
	EMsg_AMGetPlayerLinkDetailsResponse                           EMsg = 4290
	EMsg_AMGetAccountFlagsForWGSpoofing                           EMsg = 4294
	EMsg_AMGetAccountFlagsForWGSpoofingResponse                   EMsg = 4295
	EMsg_AMGetClanOfficers                                        EMsg = 4298
	EMsg_AMGetClanOfficersResponse                                EMsg = 4299
	EMsg_AMNameChange                                             EMsg = 4300
	EMsg_AMGetNameHistory                                         EMsg = 4301
	EMsg_AMGetNameHistoryResponse                                 EMsg = 4302
	EMsg_AMUpdateProviderStatus                                   EMsg = 4305
	EMsg_AMSupportRemoveAccountSecurity                           EMsg = 4307
	EMsg_AMIsAccountInCaptchaGracePeriod                          EMsg = 4308
	EMsg_AMIsAccountInCaptchaGracePeriodResponse                  EMsg = 4309
	EMsg_AMAccountPS3Unlink                                       EMsg = 4310
	EMsg_AMAccountPS3UnlinkResponse                               EMsg = 4311
	EMsg_AMStoreUserStatsResponse                                 EMsg = 4312 // Deprecated: renamed to UGSStoreUserStatsResponse
	EMsg_UGSStoreUserStatsResponse                                EMsg = 4312
	EMsg_AMGetAccountPSNInfo                                      EMsg = 4313
	EMsg_AMGetAccountPSNInfoResponse                              EMsg = 4314
	EMsg_AMAuthenticatedPlayerList                                EMsg = 4315
	EMsg_AMGetUserGifts                                           EMsg = 4316
	EMsg_AMGetUserGiftsResponse                                   EMsg = 4317
	EMsg_AMTransferLockedGifts                                    EMsg = 4320
	EMsg_AMTransferLockedGiftsResponse                            EMsg = 4321
	EMsg_AMPlayerHostedOnGameServer                               EMsg = 4322
	EMsg_AMGetAccountBanInfo                                      EMsg = 4323
	EMsg_AMGetAccountBanInfoResponse                              EMsg = 4324
	EMsg_AMRecordBanEnforcement                                   EMsg = 4325
	EMsg_AMRollbackGiftTransfer                                   EMsg = 4326
	EMsg_AMRollbackGiftTransferResponse                           EMsg = 4327
	EMsg_AMHandlePendingTransaction                               EMsg = 4328
	EMsg_AMRequestClanDetails                                     EMsg = 4329
	EMsg_AMDeleteStoredPaypalAgreement                            EMsg = 4330
	EMsg_AMGameServerUpdate                                       EMsg = 4331
	EMsg_AMGameServerRemove                                       EMsg = 4332
	EMsg_AMGetPaypalAgreements                                    EMsg = 4333
	EMsg_AMGetPaypalAgreementsResponse                            EMsg = 4334
	EMsg_AMGameServerPlayerCompatibilityCheck                     EMsg = 4335
	EMsg_AMGameServerPlayerCompatibilityCheckResponse             EMsg = 4336
	EMsg_AMRenewLicense                                           EMsg = 4337
	EMsg_AMGetAccountCommunityBanInfo                             EMsg = 4338
	EMsg_AMGetAccountCommunityBanInfoResponse                     EMsg = 4339
	EMsg_AMGameServerAccountChangePassword                        EMsg = 4340
	EMsg_AMGameServerAccountDeleteAccount                         EMsg = 4341
	EMsg_AMRenewAgreement                                         EMsg = 4342
	EMsg_AMXsollaPayment                                          EMsg = 4344
	EMsg_AMXsollaPaymentResponse                                  EMsg = 4345
	EMsg_AMAcctAllowedToPurchase                                  EMsg = 4346
	EMsg_AMAcctAllowedToPurchaseResponse                          EMsg = 4347
	EMsg_AMSwapKioskDeposit                                       EMsg = 4348
	EMsg_AMSwapKioskDepositResponse                               EMsg = 4349
	EMsg_AMSetUserGiftUnowned                                     EMsg = 4350
	EMsg_AMSetUserGiftUnownedResponse                             EMsg = 4351
	EMsg_AMClaimUnownedUserGift                                   EMsg = 4352
	EMsg_AMClaimUnownedUserGiftResponse                           EMsg = 4353
	EMsg_AMSetClanName                                            EMsg = 4354
	EMsg_AMSetClanNameResponse                                    EMsg = 4355
	EMsg_AMGrantCoupon                                            EMsg = 4356
	EMsg_AMGrantCouponResponse                                    EMsg = 4357
	EMsg_AMIsPackageRestrictedInUserCountry                       EMsg = 4358
	EMsg_AMIsPackageRestrictedInUserCountryResponse               EMsg = 4359
	EMsg_AMHandlePendingTransactionResponse                       EMsg = 4360
	EMsg_AMGrantGuestPasses2                                      EMsg = 4361
	EMsg_AMGrantGuestPasses2Response                              EMsg = 4362
	EMsg_AMSessionQuery                                           EMsg = 4363 // Deprecated
	EMsg_AMSessionQueryResponse                                   EMsg = 4364 // Deprecated
	EMsg_AMGetPlayerBanDetails                                    EMsg = 4365
	EMsg_AMGetPlayerBanDetailsResponse                            EMsg = 4366
	EMsg_AMFinalizePurchase                                       EMsg = 4367
	EMsg_AMFinalizePurchaseResponse                               EMsg = 4368
	EMsg_AMPersonaChangeResponse                                  EMsg = 4372
	EMsg_AMGetClanDetailsForForumCreation                         EMsg = 4373
	EMsg_AMGetClanDetailsForForumCreationResponse                 EMsg = 4374
	EMsg_AMGetPendingNotificationCount                            EMsg = 4375
	EMsg_AMGetPendingNotificationCountResponse                    EMsg = 4376
	EMsg_AMPasswordHashUpgrade                                    EMsg = 4377
	EMsg_AMMoPayPayment                                           EMsg = 4378 // Deprecated
	EMsg_AMMoPayPaymentResponse                                   EMsg = 4379 // Deprecated
	EMsg_AMBoaCompraPayment                                       EMsg = 4380
	EMsg_AMBoaCompraPaymentResponse                               EMsg = 4381
	EMsg_AMExpireCaptchaByGID                                     EMsg = 4382 // Deprecated
	EMsg_AMCompleteExternalPurchase                               EMsg = 4383
	EMsg_AMCompleteExternalPurchaseResponse                       EMsg = 4384
	EMsg_AMResolveNegativeWalletCredits                           EMsg = 4385
	EMsg_AMResolveNegativeWalletCreditsResponse                   EMsg = 4386
	EMsg_AMPayelpPayment                                          EMsg = 4387 // Deprecated
	EMsg_AMPayelpPaymentResponse                                  EMsg = 4388 // Deprecated
	EMsg_AMPlayerGetClanBasicDetails                              EMsg = 4389
	EMsg_AMPlayerGetClanBasicDetailsResponse                      EMsg = 4390
	EMsg_AMMOLPayment                                             EMsg = 4391
	EMsg_AMMOLPaymentResponse                                     EMsg = 4392
	EMsg_GetUserIPCountry                                         EMsg = 4393
	EMsg_GetUserIPCountryResponse                                 EMsg = 4394
	EMsg_NotificationOfSuspiciousActivity                         EMsg = 4395
	EMsg_AMDegicaPayment                                          EMsg = 4396
	EMsg_AMDegicaPaymentResponse                                  EMsg = 4397
	EMsg_AMEClubPayment                                           EMsg = 4398
	EMsg_AMEClubPaymentResponse                                   EMsg = 4399
	EMsg_AMPayPalPaymentsHubPayment                               EMsg = 4400
	EMsg_AMPayPalPaymentsHubPaymentResponse                       EMsg = 4401
	EMsg_AMTwoFactorRecoverAuthenticatorRequest                   EMsg = 4402
	EMsg_AMTwoFactorRecoverAuthenticatorResponse                  EMsg = 4403
	EMsg_AMSmart2PayPayment                                       EMsg = 4404
	EMsg_AMSmart2PayPaymentResponse                               EMsg = 4405
	EMsg_AMValidatePasswordResetCodeAndSendSmsRequest             EMsg = 4406
	EMsg_AMValidatePasswordResetCodeAndSendSmsResponse            EMsg = 4407
	EMsg_AMGetAccountResetDetailsRequest                          EMsg = 4408
	EMsg_AMGetAccountResetDetailsResponse                         EMsg = 4409
	EMsg_AMBitPayPayment                                          EMsg = 4410
	EMsg_AMBitPayPaymentResponse                                  EMsg = 4411
	EMsg_AMSendAccountInfoUpdate                                  EMsg = 4412
	EMsg_AMSendScheduledGift                                      EMsg = 4413
	EMsg_AMNodwinPayment                                          EMsg = 4414
	EMsg_AMNodwinPaymentResponse                                  EMsg = 4415
	EMsg_AMResolveWalletRevoke                                    EMsg = 4416
	EMsg_AMResolveWalletReverseRevoke                             EMsg = 4417
	EMsg_AMFundedPayment                                          EMsg = 4418
	EMsg_AMFundedPaymentResponse                                  EMsg = 4419
	EMsg_AMRequestPersonaUpdateForChatServer                      EMsg = 4420
	EMsg_AMPerfectWorldPayment                                    EMsg = 4421
	EMsg_AMPerfectWorldPaymentResponse                            EMsg = 4422
	EMsg_BasePSRange                                              EMsg = 5000
	EMsg_PSCreateShoppingCart                                     EMsg = 5001
	EMsg_PSCreateShoppingCartResponse                             EMsg = 5002
	EMsg_PSIsValidShoppingCart                                    EMsg = 5003
	EMsg_PSIsValidShoppingCartResponse                            EMsg = 5004
	EMsg_PSAddPackageToShoppingCart                               EMsg = 5005
	EMsg_PSAddPackageToShoppingCartResponse                       EMsg = 5006
	EMsg_PSRemoveLineItemFromShoppingCart                         EMsg = 5007
	EMsg_PSRemoveLineItemFromShoppingCartResponse                 EMsg = 5008
	EMsg_PSGetShoppingCartContents                                EMsg = 5009
	EMsg_PSGetShoppingCartContentsResponse                        EMsg = 5010
	EMsg_PSAddWalletCreditToShoppingCart                          EMsg = 5011
	EMsg_PSAddWalletCreditToShoppingCartResponse                  EMsg = 5012
	EMsg_BaseUFSRange                                             EMsg = 5200
	EMsg_ClientUFSUploadFileRequest                               EMsg = 5202
	EMsg_ClientUFSUploadFileResponse                              EMsg = 5203
	EMsg_ClientUFSUploadFileChunk                                 EMsg = 5204
	EMsg_ClientUFSUploadFileFinished                              EMsg = 5205
	EMsg_ClientUFSGetFileListForApp                               EMsg = 5206
	EMsg_ClientUFSGetFileListForAppResponse                       EMsg = 5207
	EMsg_ClientUFSDownloadRequest                                 EMsg = 5210
	EMsg_ClientUFSDownloadResponse                                EMsg = 5211
	EMsg_ClientUFSDownloadChunk                                   EMsg = 5212
	EMsg_ClientUFSLoginRequest                                    EMsg = 5213
	EMsg_ClientUFSLoginResponse                                   EMsg = 5214
	EMsg_UFSReloadPartitionInfo                                   EMsg = 5215
	EMsg_ClientUFSTransferHeartbeat                               EMsg = 5216
	EMsg_UFSSynchronizeFile                                       EMsg = 5217
	EMsg_UFSSynchronizeFileResponse                               EMsg = 5218
	EMsg_ClientUFSDeleteFileRequest                               EMsg = 5219
	EMsg_ClientUFSDeleteFileResponse                              EMsg = 5220
	EMsg_ClientUFSGetUGCDetails                                   EMsg = 5226
	EMsg_ClientUFSGetUGCDetailsResponse                           EMsg = 5227
	EMsg_UFSUpdateFileFlags                                       EMsg = 5228
	EMsg_UFSUpdateFileFlagsResponse                               EMsg = 5229
	EMsg_ClientUFSGetSingleFileInfo                               EMsg = 5230
	EMsg_ClientUFSGetSingleFileInfoResponse                       EMsg = 5231
	EMsg_ClientUFSShareFile                                       EMsg = 5232
	EMsg_ClientUFSShareFileResponse                               EMsg = 5233
	EMsg_UFSReloadAccount                                         EMsg = 5234
	EMsg_UFSReloadAccountResponse                                 EMsg = 5235
	EMsg_UFSUpdateRecordBatched                                   EMsg = 5236
	EMsg_UFSUpdateRecordBatchedResponse                           EMsg = 5237
	EMsg_UFSMigrateFile                                           EMsg = 5238
	EMsg_UFSMigrateFileResponse                                   EMsg = 5239
	EMsg_UFSGetUGCURLs                                            EMsg = 5240
	EMsg_UFSGetUGCURLsResponse                                    EMsg = 5241
	EMsg_UFSHttpUploadFileFinishRequest                           EMsg = 5242
	EMsg_UFSHttpUploadFileFinishResponse                          EMsg = 5243
	EMsg_UFSDownloadStartRequest                                  EMsg = 5244
	EMsg_UFSDownloadStartResponse                                 EMsg = 5245
	EMsg_UFSDownloadChunkRequest                                  EMsg = 5246
	EMsg_UFSDownloadChunkResponse                                 EMsg = 5247
	EMsg_UFSDownloadFinishRequest                                 EMsg = 5248
	EMsg_UFSDownloadFinishResponse                                EMsg = 5249
	EMsg_UFSFlushURLCache                                         EMsg = 5250
	EMsg_UFSUploadCommit                                          EMsg = 5251 // Deprecated: renamed to ClientUFSUploadCommit
	EMsg_ClientUFSUploadCommit                                    EMsg = 5251
	EMsg_UFSUploadCommitResponse                                  EMsg = 5252 // Deprecated: renamed to ClientUFSUploadCommitResponse
	EMsg_ClientUFSUploadCommitResponse                            EMsg = 5252
	EMsg_UFSMigrateFileAppID                                      EMsg = 5253
	EMsg_UFSMigrateFileAppIDResponse                              EMsg = 5254
	EMsg_BaseClient2                                              EMsg = 5400
	EMsg_ClientRequestForgottenPasswordEmail                      EMsg = 5401
	EMsg_ClientRequestForgottenPasswordEmailResponse              EMsg = 5402
	EMsg_ClientCreateAccountResponse                              EMsg = 5403
	EMsg_ClientResetForgottenPassword                             EMsg = 5404
	EMsg_ClientResetForgottenPasswordResponse                     EMsg = 5405
	EMsg_ClientCreateAccount2                                     EMsg = 5406 // Deprecated
	EMsg_ClientInformOfResetForgottenPassword                     EMsg = 5407
	EMsg_ClientInformOfResetForgottenPasswordResponse             EMsg = 5408
	EMsg_ClientGamesPlayedWithDataBlob                            EMsg = 5410
	EMsg_ClientUpdateUserGameInfo                                 EMsg = 5411
	EMsg_ClientFileToDownload                                     EMsg = 5412
	EMsg_ClientFileToDownloadResponse                             EMsg = 5413
	EMsg_ClientLBSSetScore                                        EMsg = 5414
	EMsg_ClientLBSSetScoreResponse                                EMsg = 5415
	EMsg_ClientLBSFindOrCreateLB                                  EMsg = 5416
	EMsg_ClientLBSFindOrCreateLBResponse                          EMsg = 5417
	EMsg_ClientLBSGetLBEntries                                    EMsg = 5418
	EMsg_ClientLBSGetLBEntriesResponse                            EMsg = 5419
	EMsg_ClientChatDeclined                                       EMsg = 5426
	EMsg_ClientFriendMsgIncoming                                  EMsg = 5427
	EMsg_ClientTicketAuthComplete                                 EMsg = 5429
	EMsg_ClientIsLimitedAccount                                   EMsg = 5430
	EMsg_ClientRequestAuthList                                    EMsg = 5431
	EMsg_ClientAuthList                                           EMsg = 5432
	EMsg_ClientStat                                               EMsg = 5433
	EMsg_ClientP2PConnectionInfo                                  EMsg = 5434
	EMsg_ClientP2PConnectionFailInfo                              EMsg = 5435
	EMsg_ClientGetDepotDecryptionKey                              EMsg = 5438
	EMsg_ClientGetDepotDecryptionKeyResponse                      EMsg = 5439
	EMsg_GSPerformHardwareSurvey                                  EMsg = 5440
	EMsg_ClientEnableTestLicense                                  EMsg = 5443
	EMsg_ClientEnableTestLicenseResponse                          EMsg = 5444
	EMsg_ClientDisableTestLicense                                 EMsg = 5445
	EMsg_ClientDisableTestLicenseResponse                         EMsg = 5446
	EMsg_ClientRequestValidationMail                              EMsg = 5448
	EMsg_ClientRequestValidationMailResponse                      EMsg = 5449
	EMsg_ClientCheckAppBetaPassword                               EMsg = 5450
	EMsg_ClientCheckAppBetaPasswordResponse                       EMsg = 5451
	EMsg_ClientToGC                                               EMsg = 5452
	EMsg_ClientFromGC                                             EMsg = 5453
	EMsg_ClientRequestChangeMail                                  EMsg = 5454
	EMsg_ClientRequestChangeMailResponse                          EMsg = 5455
	EMsg_ClientEmailAddrInfo                                      EMsg = 5456
	EMsg_ClientPasswordChange3                                    EMsg = 5457
	EMsg_ClientEmailChange3                                       EMsg = 5458
	EMsg_ClientPersonalQAChange3                                  EMsg = 5459
	EMsg_ClientResetForgottenPassword3                            EMsg = 5460
	EMsg_ClientRequestForgottenPasswordEmail3                     EMsg = 5461
	EMsg_ClientNewLoginKey                                        EMsg = 5463
	EMsg_ClientNewLoginKeyAccepted                                EMsg = 5464
	EMsg_ClientStoreUserStats2                                    EMsg = 5466
	EMsg_ClientStatsUpdated                                       EMsg = 5467
	EMsg_ClientActivateOEMLicense                                 EMsg = 5468
	EMsg_ClientRegisterOEMMachine                                 EMsg = 5469
	EMsg_ClientRegisterOEMMachineResponse                         EMsg = 5470
	EMsg_ClientRequestedClientStats                               EMsg = 5480
	EMsg_ClientStat2Int32                                         EMsg = 5481
	EMsg_ClientStat2                                              EMsg = 5482
	EMsg_ClientVerifyPassword                                     EMsg = 5483
	EMsg_ClientVerifyPasswordResponse                             EMsg = 5484
	EMsg_ClientDRMDownloadRequest                                 EMsg = 5485
	EMsg_ClientDRMDownloadResponse                                EMsg = 5486
	EMsg_ClientDRMFinalResult                                     EMsg = 5487
	EMsg_ClientGetFriendsWhoPlayGame                              EMsg = 5488
	EMsg_ClientGetFriendsWhoPlayGameResponse                      EMsg = 5489
	EMsg_ClientOGSBeginSession                                    EMsg = 5490
	EMsg_ClientOGSBeginSessionResponse                            EMsg = 5491
	EMsg_ClientOGSEndSession                                      EMsg = 5492
	EMsg_ClientOGSEndSessionResponse                              EMsg = 5493
	EMsg_ClientOGSWriteRow                                        EMsg = 5494
	EMsg_ClientDRMTest                                            EMsg = 5495
	EMsg_ClientDRMTestResult                                      EMsg = 5496
	EMsg_ClientServerUnavailable                                  EMsg = 5500
	EMsg_ClientServersAvailable                                   EMsg = 5501
	EMsg_ClientRegisterAuthTicketWithCM                           EMsg = 5502
	EMsg_ClientGCMsgFailed                                        EMsg = 5503
	EMsg_ClientMicroTxnAuthRequest                                EMsg = 5504
	EMsg_ClientMicroTxnAuthorize                                  EMsg = 5505
	EMsg_ClientMicroTxnAuthorizeResponse                          EMsg = 5506
	EMsg_ClientAppMinutesPlayedData                               EMsg = 5507
	EMsg_ClientGetMicroTxnInfo                                    EMsg = 5508
	EMsg_ClientGetMicroTxnInfoResponse                            EMsg = 5509
	EMsg_ClientMarketingMessageUpdate2                            EMsg = 5510
	EMsg_ClientDeregisterWithServer                               EMsg = 5511
	EMsg_ClientSubscribeToPersonaFeed                             EMsg = 5512
	EMsg_ClientLogon                                              EMsg = 5514
	EMsg_ClientGetClientDetails                                   EMsg = 5515
	EMsg_ClientGetClientDetailsResponse                           EMsg = 5516
	EMsg_ClientReportOverlayDetourFailure                         EMsg = 5517
	EMsg_ClientGetClientAppList                                   EMsg = 5518
	EMsg_ClientGetClientAppListResponse                           EMsg = 5519
	EMsg_ClientInstallClientApp                                   EMsg = 5520
	EMsg_ClientInstallClientAppResponse                           EMsg = 5521
	EMsg_ClientUninstallClientApp                                 EMsg = 5522
	EMsg_ClientUninstallClientAppResponse                         EMsg = 5523
	EMsg_ClientSetClientAppUpdateState                            EMsg = 5524
	EMsg_ClientSetClientAppUpdateStateResponse                    EMsg = 5525
	EMsg_ClientRequestEncryptedAppTicket                          EMsg = 5526
	EMsg_ClientRequestEncryptedAppTicketResponse                  EMsg = 5527
	EMsg_ClientWalletInfoUpdate                                   EMsg = 5528
	EMsg_ClientLBSSetUGC                                          EMsg = 5529
	EMsg_ClientLBSSetUGCResponse                                  EMsg = 5530
	EMsg_ClientAMGetClanOfficers                                  EMsg = 5531
	EMsg_ClientAMGetClanOfficersResponse                          EMsg = 5532
	EMsg_ClientFriendProfileInfo                                  EMsg = 5535
	EMsg_ClientFriendProfileInfoResponse                          EMsg = 5536
	EMsg_ClientUpdateMachineAuth                                  EMsg = 5537
	EMsg_ClientUpdateMachineAuthResponse                          EMsg = 5538
	EMsg_ClientReadMachineAuth                                    EMsg = 5539
	EMsg_ClientReadMachineAuthResponse                            EMsg = 5540
	EMsg_ClientRequestMachineAuth                                 EMsg = 5541
	EMsg_ClientRequestMachineAuthResponse                         EMsg = 5542
	EMsg_ClientScreenshotsChanged                                 EMsg = 5543
	EMsg_ClientEmailChange4                                       EMsg = 5544 // Deprecated
	EMsg_ClientEmailChangeResponse4                               EMsg = 5545 // Deprecated
	EMsg_ClientGetCDNAuthToken                                    EMsg = 5546
	EMsg_ClientGetCDNAuthTokenResponse                            EMsg = 5547
	EMsg_ClientDownloadRateStatistics                             EMsg = 5548
	EMsg_ClientRequestAccountData                                 EMsg = 5549
	EMsg_ClientRequestAccountDataResponse                         EMsg = 5550
	EMsg_ClientResetForgottenPassword4                            EMsg = 5551
	EMsg_ClientHideFriend                                         EMsg = 5552
	EMsg_ClientFriendsGroupsList                                  EMsg = 5553
	EMsg_ClientGetClanActivityCounts                              EMsg = 5554
	EMsg_ClientGetClanActivityCountsResponse                      EMsg = 5555
	EMsg_ClientOGSReportString                                    EMsg = 5556
	EMsg_ClientOGSReportBug                                       EMsg = 5557
	EMsg_ClientSentLogs                                           EMsg = 5558
	EMsg_ClientLogonGameServer                                    EMsg = 5559
	EMsg_AMClientCreateFriendsGroup                               EMsg = 5560
	EMsg_AMClientCreateFriendsGroupResponse                       EMsg = 5561
	EMsg_AMClientDeleteFriendsGroup                               EMsg = 5562
	EMsg_AMClientDeleteFriendsGroupResponse                       EMsg = 5563
	EMsg_AMClientRenameFriendsGroup                               EMsg = 5564 // Deprecated: renamed to AMClientManageFriendsGroup
	EMsg_AMClientManageFriendsGroup                               EMsg = 5564
	EMsg_AMClientRenameFriendsGroupResponse                       EMsg = 5565 // Deprecated: renamed to AMClientManageFriendsGroupResponse
	EMsg_AMClientManageFriendsGroupResponse                       EMsg = 5565
	EMsg_AMClientAddFriendToGroup                                 EMsg = 5566
	EMsg_AMClientAddFriendToGroupResponse                         EMsg = 5567
	EMsg_AMClientRemoveFriendFromGroup                            EMsg = 5568
	EMsg_AMClientRemoveFriendFromGroupResponse                    EMsg = 5569
	EMsg_ClientAMGetPersonaNameHistory                            EMsg = 5570
	EMsg_ClientAMGetPersonaNameHistoryResponse                    EMsg = 5571
	EMsg_ClientRequestFreeLicense                                 EMsg = 5572
	EMsg_ClientRequestFreeLicenseResponse                         EMsg = 5573
	EMsg_ClientDRMDownloadRequestWithCrashData                    EMsg = 5574
	EMsg_ClientAuthListAck                                        EMsg = 5575
	EMsg_ClientItemAnnouncements                                  EMsg = 5576
	EMsg_ClientRequestItemAnnouncements                           EMsg = 5577
	EMsg_ClientFriendMsgEchoToSender                              EMsg = 5578
	EMsg_ClientOGSGameServerPingSample                            EMsg = 5581
	EMsg_ClientCommentNotifications                               EMsg = 5582
	EMsg_ClientRequestCommentNotifications                        EMsg = 5583
	EMsg_ClientPersonaChangeResponse                              EMsg = 5584
	EMsg_ClientRequestWebAPIAuthenticateUserNonce                 EMsg = 5585
	EMsg_ClientRequestWebAPIAuthenticateUserNonceResponse         EMsg = 5586
	EMsg_ClientPlayerNicknameList                                 EMsg = 5587
	EMsg_AMClientSetPlayerNickname                                EMsg = 5588
	EMsg_AMClientSetPlayerNicknameResponse                        EMsg = 5589
	EMsg_ClientCreateAccountProto                                 EMsg = 5590 // Deprecated
	EMsg_ClientCreateAccountProtoResponse                         EMsg = 5591 // Deprecated
	EMsg_ClientGetNumberOfCurrentPlayersDP                        EMsg = 5592
	EMsg_ClientGetNumberOfCurrentPlayersDPResponse                EMsg = 5593
	EMsg_ClientServiceMethod                                      EMsg = 5594 // Deprecated: renamed to ClientServiceMethodLegacy
	EMsg_ClientServiceMethodLegacy                                EMsg = 5594
	EMsg_ClientServiceMethodResponse                              EMsg = 5595 // Deprecated: renamed to ClientServiceMethodLegacyResponse
	EMsg_ClientServiceMethodLegacyResponse                        EMsg = 5595
	EMsg_ClientFriendUserStatusPublished                          EMsg = 5596
	EMsg_ClientCurrentUIMode                                      EMsg = 5597
	EMsg_ClientVanityURLChangedNotification                       EMsg = 5598
	EMsg_ClientUserNotifications                                  EMsg = 5599
	EMsg_BaseDFS                                                  EMsg = 5600
	EMsg_DFSGetFile                                               EMsg = 5601
	EMsg_DFSInstallLocalFile                                      EMsg = 5602
	EMsg_DFSConnection                                            EMsg = 5603
	EMsg_DFSConnectionReply                                       EMsg = 5604
	EMsg_ClientDFSAuthenticateRequest                             EMsg = 5605
	EMsg_ClientDFSAuthenticateResponse                            EMsg = 5606
	EMsg_ClientDFSEndSession                                      EMsg = 5607
	EMsg_DFSPurgeFile                                             EMsg = 5608
	EMsg_DFSRouteFile                                             EMsg = 5609
	EMsg_DFSGetFileFromServer                                     EMsg = 5610
	EMsg_DFSAcceptedResponse                                      EMsg = 5611
	EMsg_DFSRequestPingback                                       EMsg = 5612
	EMsg_DFSRecvTransmitFile                                      EMsg = 5613
	EMsg_DFSSendTransmitFile                                      EMsg = 5614
	EMsg_DFSRequestPingback2                                      EMsg = 5615
	EMsg_DFSResponsePingback2                                     EMsg = 5616
	EMsg_ClientDFSDownloadStatus                                  EMsg = 5617
	EMsg_DFSStartTransfer                                         EMsg = 5618
	EMsg_DFSTransferComplete                                      EMsg = 5619
	EMsg_DFSRouteFileResponse                                     EMsg = 5620
	EMsg_ClientNetworkingCertRequest                              EMsg = 5621
	EMsg_ClientNetworkingCertRequestResponse                      EMsg = 5622
	EMsg_ClientChallengeRequest                                   EMsg = 5623
	EMsg_ClientChallengeResponse                                  EMsg = 5624
	EMsg_BadgeCraftedNotification                                 EMsg = 5625
	EMsg_ClientNetworkingMobileCertRequest                        EMsg = 5626
	EMsg_ClientNetworkingMobileCertRequestResponse                EMsg = 5627
	EMsg_BaseMDS                                                  EMsg = 5800
	EMsg_AMToMDSGetDepotDecryptionKey                             EMsg = 5812
	EMsg_MDSToAMGetDepotDecryptionKeyResponse                     EMsg = 5813
	EMsg_MDSContentServerConfigRequest                            EMsg = 5827
	EMsg_MDSContentServerConfig                                   EMsg = 5828
	EMsg_MDSGetDepotManifest                                      EMsg = 5829
	EMsg_MDSGetDepotManifestResponse                              EMsg = 5830
	EMsg_MDSGetDepotManifestChunk                                 EMsg = 5831
	EMsg_MDSGetDepotChunk                                         EMsg = 5832
	EMsg_MDSGetDepotChunkResponse                                 EMsg = 5833
	EMsg_MDSGetDepotChunkChunk                                    EMsg = 5834
	EMsg_MDSGetServerListForUser                                  EMsg = 5836 // Deprecated
	EMsg_MDSGetServerListForUserResponse                          EMsg = 5837 // Deprecated
	EMsg_MDSToCSFlushChunk                                        EMsg = 5844
	EMsg_MDSMigrateChunk                                          EMsg = 5847
	EMsg_MDSMigrateChunkResponse                                  EMsg = 5848
	EMsg_MDSToCSFlushManifest                                     EMsg = 5849
	EMsg_CSBase                                                   EMsg = 6200
	EMsg_CSPing                                                   EMsg = 6201
	EMsg_CSPingResponse                                           EMsg = 6202
	EMsg_GMSBase                                                  EMsg = 6400
	EMsg_GMSGameServerReplicate                                   EMsg = 6401
	EMsg_ClientGMSServerQuery                                     EMsg = 6403
	EMsg_GMSClientServerQueryResponse                             EMsg = 6404
	EMsg_AMGMSGameServerUpdate                                    EMsg = 6405
	EMsg_AMGMSGameServerRemove                                    EMsg = 6406
	EMsg_GameServerOutOfDate                                      EMsg = 6407
	EMsg_DeviceAuthorizationBase                                  EMsg = 6500
	EMsg_ClientAuthorizeLocalDeviceRequest                        EMsg = 6501
	EMsg_ClientAuthorizeLocalDeviceResponse                       EMsg = 6502
	EMsg_ClientDeauthorizeDeviceRequest                           EMsg = 6503
	EMsg_ClientDeauthorizeDevice                                  EMsg = 6504
	EMsg_ClientUseLocalDeviceAuthorizations                       EMsg = 6505
	EMsg_ClientGetAuthorizedDevices                               EMsg = 6506
	EMsg_ClientGetAuthorizedDevicesResponse                       EMsg = 6507
	EMsg_AMNotifySessionDeviceAuthorized                          EMsg = 6508
	EMsg_ClientAuthorizeLocalDeviceNotification                   EMsg = 6509
	EMsg_MMSBase                                                  EMsg = 6600
	EMsg_ClientMMSCreateLobby                                     EMsg = 6601
	EMsg_ClientMMSCreateLobbyResponse                             EMsg = 6602
	EMsg_ClientMMSJoinLobby                                       EMsg = 6603
	EMsg_ClientMMSJoinLobbyResponse                               EMsg = 6604
	EMsg_ClientMMSLeaveLobby                                      EMsg = 6605
	EMsg_ClientMMSLeaveLobbyResponse                              EMsg = 6606
	EMsg_ClientMMSGetLobbyList                                    EMsg = 6607
	EMsg_ClientMMSGetLobbyListResponse                            EMsg = 6608
	EMsg_ClientMMSSetLobbyData                                    EMsg = 6609
	EMsg_ClientMMSSetLobbyDataResponse                            EMsg = 6610
	EMsg_ClientMMSGetLobbyData                                    EMsg = 6611
	EMsg_ClientMMSLobbyData                                       EMsg = 6612
	EMsg_ClientMMSSendLobbyChatMsg                                EMsg = 6613
	EMsg_ClientMMSLobbyChatMsg                                    EMsg = 6614
	EMsg_ClientMMSSetLobbyOwner                                   EMsg = 6615
	EMsg_ClientMMSSetLobbyOwnerResponse                           EMsg = 6616
	EMsg_ClientMMSSetLobbyGameServer                              EMsg = 6617
	EMsg_ClientMMSLobbyGameServerSet                              EMsg = 6618
	EMsg_ClientMMSUserJoinedLobby                                 EMsg = 6619
	EMsg_ClientMMSUserLeftLobby                                   EMsg = 6620
	EMsg_ClientMMSInviteToLobby                                   EMsg = 6621
	EMsg_ClientMMSFlushFrenemyListCache                           EMsg = 6622
	EMsg_ClientMMSFlushFrenemyListCacheResponse                   EMsg = 6623
	EMsg_ClientMMSSetLobbyLinked                                  EMsg = 6624
	EMsg_ClientMMSSetRatelimitPolicyOnClient                      EMsg = 6625
	EMsg_ClientMMSGetLobbyStatus                                  EMsg = 6626
	EMsg_ClientMMSGetLobbyStatusResponse                          EMsg = 6627
	EMsg_MMSGetLobbyList                                          EMsg = 6628
	EMsg_MMSGetLobbyListResponse                                  EMsg = 6629
	EMsg_NonStdMsgBase                                            EMsg = 6800
	EMsg_NonStdMsgMemcached                                       EMsg = 6801
	EMsg_NonStdMsgHTTPServer                                      EMsg = 6802
	EMsg_NonStdMsgHTTPClient                                      EMsg = 6803
	EMsg_NonStdMsgWGResponse                                      EMsg = 6804
	EMsg_NonStdMsgPHPSimulator                                    EMsg = 6805
	EMsg_NonStdMsgChase                                           EMsg = 6806
	EMsg_NonStdMsgDFSTransfer                                     EMsg = 6807
	EMsg_NonStdMsgTests                                           EMsg = 6808
	EMsg_NonStdMsgUMQpipeAAPL                                     EMsg = 6809
	EMsg_NonStdMsgSyslog                                          EMsg = 6810
	EMsg_NonStdMsgLogsink                                         EMsg = 6811
	EMsg_NonStdMsgSteam2Emulator                                  EMsg = 6812
	EMsg_NonStdMsgRTMPServer                                      EMsg = 6813
	EMsg_NonStdMsgWebSocket                                       EMsg = 6814
	EMsg_NonStdMsgRedis                                           EMsg = 6815
	EMsg_UDSBase                                                  EMsg = 7000
	EMsg_ClientUDSP2PSessionStarted                               EMsg = 7001
	EMsg_ClientUDSP2PSessionEnded                                 EMsg = 7002
	EMsg_UDSRenderUserAuth                                        EMsg = 7003
	EMsg_UDSRenderUserAuthResponse                                EMsg = 7004
	EMsg_ClientUDSInviteToGame                                    EMsg = 7005 // Deprecated: renamed to ClientInviteToGame
	EMsg_ClientInviteToGame                                       EMsg = 7005
	EMsg_UDSHasSession                                            EMsg = 7006
	EMsg_UDSHasSessionResponse                                    EMsg = 7007
	EMsg_MPASBase                                                 EMsg = 7100
	EMsg_MPASVacBanReset                                          EMsg = 7101
	EMsg_KGSBase                                                  EMsg = 7200
	EMsg_UCMBase                                                  EMsg = 7300
	EMsg_ClientUCMAddScreenshot                                   EMsg = 7301
	EMsg_ClientUCMAddScreenshotResponse                           EMsg = 7302
	EMsg_UCMResetCommunityContent                                 EMsg = 7307
	EMsg_UCMResetCommunityContentResponse                         EMsg = 7308
	EMsg_ClientUCMDeleteScreenshot                                EMsg = 7309
	EMsg_ClientUCMDeleteScreenshotResponse                        EMsg = 7310
	EMsg_ClientUCMPublishFile                                     EMsg = 7311
	EMsg_ClientUCMPublishFileResponse                             EMsg = 7312
	EMsg_ClientUCMDeletePublishedFile                             EMsg = 7315
	EMsg_ClientUCMDeletePublishedFileResponse                     EMsg = 7316
	EMsg_ClientUCMEnumerateUserPublishedFiles                     EMsg = 7317
	EMsg_ClientUCMEnumerateUserPublishedFilesResponse             EMsg = 7318
	EMsg_ClientUCMEnumerateUserSubscribedFiles                    EMsg = 7321
	EMsg_ClientUCMEnumerateUserSubscribedFilesResponse            EMsg = 7322
	EMsg_ClientUCMUpdatePublishedFile                             EMsg = 7325
	EMsg_ClientUCMUpdatePublishedFileResponse                     EMsg = 7326
	EMsg_UCMUpdatePublishedFile                                   EMsg = 7327
	EMsg_UCMUpdatePublishedFileResponse                           EMsg = 7328
	EMsg_UCMDeletePublishedFile                                   EMsg = 7329
	EMsg_UCMDeletePublishedFileResponse                           EMsg = 7330
	EMsg_UCMUpdatePublishedFileStat                               EMsg = 7331
	EMsg_UCMUpdatePublishedFileBan                                EMsg = 7332 // Deprecated
	EMsg_UCMUpdatePublishedFileBanResponse                        EMsg = 7333 // Deprecated
	EMsg_UCMReloadPublishedFile                                   EMsg = 7337
	EMsg_UCMReloadUserFileListCaches                              EMsg = 7338
	EMsg_UCMPublishedFileReported                                 EMsg = 7339
	EMsg_UCMUpdatePublishedFileIncompatibleStatus                 EMsg = 7340 // Deprecated
	EMsg_UCMPublishedFilePreviewAdd                               EMsg = 7341
	EMsg_UCMPublishedFilePreviewAddResponse                       EMsg = 7342
	EMsg_UCMPublishedFilePreviewRemove                            EMsg = 7343
	EMsg_UCMPublishedFilePreviewRemoveResponse                    EMsg = 7344
	EMsg_ClientUCMPublishedFileSubscribed                         EMsg = 7347
	EMsg_ClientUCMPublishedFileUnsubscribed                       EMsg = 7348
	EMsg_UCMPublishedFileSubscribed                               EMsg = 7349
	EMsg_UCMPublishedFileUnsubscribed                             EMsg = 7350
	EMsg_UCMPublishFile                                           EMsg = 7351
	EMsg_UCMPublishFileResponse                                   EMsg = 7352
	EMsg_UCMPublishedFileChildAdd                                 EMsg = 7353
	EMsg_UCMPublishedFileChildAddResponse                         EMsg = 7354
	EMsg_UCMPublishedFileChildRemove                              EMsg = 7355
	EMsg_UCMPublishedFileChildRemoveResponse                      EMsg = 7356
	EMsg_UCMPublishedFileParentChanged                            EMsg = 7359
	EMsg_ClientUCMGetPublishedFilesForUser                        EMsg = 7360
	EMsg_ClientUCMGetPublishedFilesForUserResponse                EMsg = 7361
	EMsg_ClientUCMSetUserPublishedFileAction                      EMsg = 7364
	EMsg_ClientUCMSetUserPublishedFileActionResponse              EMsg = 7365
	EMsg_ClientUCMEnumeratePublishedFilesByUserAction             EMsg = 7366
	EMsg_ClientUCMEnumeratePublishedFilesByUserActionResponse     EMsg = 7367
	EMsg_ClientUCMPublishedFileDeleted                            EMsg = 7368
	EMsg_UCMGetUserSubscribedFiles                                EMsg = 7369
	EMsg_UCMGetUserSubscribedFilesResponse                        EMsg = 7370
	EMsg_UCMFixStatsPublishedFile                                 EMsg = 7371
	EMsg_ClientUCMEnumerateUserSubscribedFilesWithUpdates         EMsg = 7378
	EMsg_ClientUCMEnumerateUserSubscribedFilesWithUpdatesResponse EMsg = 7379
	EMsg_UCMPublishedFileContentUpdated                           EMsg = 7380
	EMsg_UCMPublishedFileUpdated                                  EMsg = 7381 // Deprecated: renamed to ClientUCMPublishedFileUpdated
	EMsg_ClientUCMPublishedFileUpdated                            EMsg = 7381
	EMsg_ClientWorkshopItemChangesRequest                         EMsg = 7382
	EMsg_ClientWorkshopItemChangesResponse                        EMsg = 7383
	EMsg_ClientWorkshopItemInfoRequest                            EMsg = 7384
	EMsg_ClientWorkshopItemInfoResponse                           EMsg = 7385
	EMsg_FSBase                                                   EMsg = 7500
	EMsg_ClientRichPresenceUpload                                 EMsg = 7501
	EMsg_ClientRichPresenceRequest                                EMsg = 7502
	EMsg_ClientRichPresenceInfo                                   EMsg = 7503
	EMsg_FSRichPresenceRequest                                    EMsg = 7504
	EMsg_FSRichPresenceResponse                                   EMsg = 7505
	EMsg_FSComputeFrenematrix                                     EMsg = 7506
	EMsg_FSComputeFrenematrixResponse                             EMsg = 7507
	EMsg_FSPlayStatusNotification                                 EMsg = 7508
	EMsg_FSPublishPersonaStatus                                   EMsg = 7509 // Deprecated
	EMsg_FSAddOrRemoveFollower                                    EMsg = 7510
	EMsg_FSAddOrRemoveFollowerResponse                            EMsg = 7511
	EMsg_FSUpdateFollowingList                                    EMsg = 7512
	EMsg_FSCommentNotification                                    EMsg = 7513
	EMsg_FSCommentNotificationViewed                              EMsg = 7514
	EMsg_ClientFSGetFollowerCount                                 EMsg = 7515
	EMsg_ClientFSGetFollowerCountResponse                         EMsg = 7516
	EMsg_ClientFSGetIsFollowing                                   EMsg = 7517
	EMsg_ClientFSGetIsFollowingResponse                           EMsg = 7518
	EMsg_ClientFSEnumerateFollowingList                           EMsg = 7519
	EMsg_ClientFSEnumerateFollowingListResponse                   EMsg = 7520
	EMsg_FSGetPendingNotificationCount                            EMsg = 7521
	EMsg_FSGetPendingNotificationCountResponse                    EMsg = 7522
	EMsg_ClientFSOfflineMessageNotification                       EMsg = 7523 // Deprecated: Renamed to ClientChatOfflineMessageNotification
	EMsg_ClientFSRequestOfflineMessageCount                       EMsg = 7524 // Deprecated: Renamed to ClientChatRequestOfflineMessageCount
	EMsg_ClientFSGetFriendMessageHistory                          EMsg = 7525 // Deprecated: Renamed to ClientChatGetFriendMessageHistory
	EMsg_ClientFSGetFriendMessageHistoryResponse                  EMsg = 7526 // Deprecated: Renamed to ClientChatGetFriendMessageHistoryResponse
	EMsg_ClientFSGetFriendMessageHistoryForOfflineMessages        EMsg = 7527 // Deprecated: Renamed to ClientChatGetFriendMessageHistoryForOfflineMessages
	EMsg_ClientChatOfflineMessageNotification                     EMsg = 7523
	EMsg_ClientChatRequestOfflineMessageCount                     EMsg = 7524
	EMsg_ClientChatGetFriendMessageHistory                        EMsg = 7525
	EMsg_ClientChatGetFriendMessageHistoryResponse                EMsg = 7526
	EMsg_ClientChatGetFriendMessageHistoryForOfflineMessages      EMsg = 7527
	EMsg_ClientFSGetFriendsSteamLevels                            EMsg = 7528
	EMsg_ClientFSGetFriendsSteamLevelsResponse                    EMsg = 7529
	EMsg_FSRequestFriendData                                      EMsg = 7530 // Deprecated: renamed to AMRequestFriendData
	EMsg_AMRequestFriendData                                      EMsg = 7530
	EMsg_DRMRange2                                                EMsg = 7600
	EMsg_CEGVersionSetEnableDisableRequest                        EMsg = 7600
	EMsg_CEGVersionSetEnableDisableResponse                       EMsg = 7601
	EMsg_CEGPropStatusDRMSRequest                                 EMsg = 7602
	EMsg_CEGPropStatusDRMSResponse                                EMsg = 7603
	EMsg_CEGWhackFailureReportRequest                             EMsg = 7604
	EMsg_CEGWhackFailureReportResponse                            EMsg = 7605
	EMsg_DRMSFetchVersionSet                                      EMsg = 7606
	EMsg_DRMSFetchVersionSetResponse                              EMsg = 7607
	EMsg_EconBase                                                 EMsg = 7700
	EMsg_EconTrading_InitiateTradeRequest                         EMsg = 7701
	EMsg_EconTrading_InitiateTradeProposed                        EMsg = 7702
	EMsg_EconTrading_InitiateTradeResponse                        EMsg = 7703
	EMsg_EconTrading_InitiateTradeResult                          EMsg = 7704
	EMsg_EconTrading_StartSession                                 EMsg = 7705
	EMsg_EconTrading_CancelTradeRequest                           EMsg = 7706
	EMsg_EconFlushInventoryCache                                  EMsg = 7707
	EMsg_EconFlushInventoryCacheResponse                          EMsg = 7708
	EMsg_EconCDKeyProcessTransaction                              EMsg = 7711
	EMsg_EconCDKeyProcessTransactionResponse                      EMsg = 7712
	EMsg_EconGetErrorLogs                                         EMsg = 7713
	EMsg_EconGetErrorLogsResponse                                 EMsg = 7714
	EMsg_RMRange                                                  EMsg = 7800
	EMsg_RMTestVerisignOTP                                        EMsg = 7800
	EMsg_RMTestVerisignOTPResponse                                EMsg = 7801
	EMsg_RMDeleteMemcachedKeys                                    EMsg = 7803
	EMsg_RMRemoteInvoke                                           EMsg = 7804
	EMsg_BadLoginIPList                                           EMsg = 7805
	EMsg_RMMsgTraceAddTrigger                                     EMsg = 7806
	EMsg_RMMsgTraceRemoveTrigger                                  EMsg = 7807
	EMsg_RMMsgTraceEvent                                          EMsg = 7808
	EMsg_UGSBase                                                  EMsg = 7900
	EMsg_UGSUpdateGlobalStats                                     EMsg = 7900
	EMsg_ClientUGSGetGlobalStats                                  EMsg = 7901
	EMsg_ClientUGSGetGlobalStatsResponse                          EMsg = 7902
	EMsg_StoreBase                                                EMsg = 8000
	EMsg_UMQBase                                                  EMsg = 8100
	EMsg_UMQLogonRequest                                          EMsg = 8100
	EMsg_UMQLogonResponse                                         EMsg = 8101
	EMsg_UMQLogoffRequest                                         EMsg = 8102
	EMsg_UMQLogoffResponse                                        EMsg = 8103
	EMsg_UMQSendChatMessage                                       EMsg = 8104
	EMsg_UMQIncomingChatMessage                                   EMsg = 8105
	EMsg_UMQPoll                                                  EMsg = 8106
	EMsg_UMQPollResults                                           EMsg = 8107
	EMsg_UMQ2AM_ClientMsgBatch                                    EMsg = 8108
	EMsg_WorkshopBase                                             EMsg = 8200
	EMsg_WebAPIBase                                               EMsg = 8300
	EMsg_WebAPIValidateOAuth2Token                                EMsg = 8300
	EMsg_WebAPIValidateOAuth2TokenResponse                        EMsg = 8301
	EMsg_WebAPIRegisterGCInterfaces                               EMsg = 8303
	EMsg_WebAPIInvalidateOAuthClientCache                         EMsg = 8304
	EMsg_WebAPIInvalidateOAuthTokenCache                          EMsg = 8305
	EMsg_WebAPISetSecrets                                         EMsg = 8306
	EMsg_BackpackBase                                             EMsg = 8400
	EMsg_BackpackAddToCurrency                                    EMsg = 8401
	EMsg_BackpackAddToCurrencyResponse                            EMsg = 8402
	EMsg_CREBase                                                  EMsg = 8500
	EMsg_CREItemVoteSummary                                       EMsg = 8503
	EMsg_CREItemVoteSummaryResponse                               EMsg = 8504
	EMsg_CREUpdateUserPublishedItemVote                           EMsg = 8507
	EMsg_CREUpdateUserPublishedItemVoteResponse                   EMsg = 8508
	EMsg_CREGetUserPublishedItemVoteDetails                       EMsg = 8509
	EMsg_CREGetUserPublishedItemVoteDetailsResponse               EMsg = 8510
	EMsg_CREEnumeratePublishedFiles                               EMsg = 8511
	EMsg_CREEnumeratePublishedFilesResponse                       EMsg = 8512
	EMsg_CREPublishedFileVoteAdded                                EMsg = 8513
	EMsg_SecretsBase                                              EMsg = 8600
	EMsg_SecretsRequestCredentialPair                             EMsg = 8600
	EMsg_SecretsCredentialPairResponse                            EMsg = 8601
	EMsg_BoxMonitorBase                                           EMsg = 8700
	EMsg_BoxMonitorReportRequest                                  EMsg = 8700
	EMsg_BoxMonitorReportResponse                                 EMsg = 8701
	EMsg_LogsinkBase                                              EMsg = 8800
	EMsg_LogsinkWriteReport                                       EMsg = 8800
	EMsg_PICSBase                                                 EMsg = 8900
	EMsg_ClientPICSChangesSinceRequest                            EMsg = 8901
	EMsg_ClientPICSChangesSinceResponse                           EMsg = 8902
	EMsg_ClientPICSProductInfoRequest                             EMsg = 8903
	EMsg_ClientPICSProductInfoResponse                            EMsg = 8904
	EMsg_ClientPICSAccessTokenRequest                             EMsg = 8905
	EMsg_ClientPICSAccessTokenResponse                            EMsg = 8906
	EMsg_WorkerProcess                                            EMsg = 9000
	EMsg_WorkerProcessPingRequest                                 EMsg = 9000
	EMsg_WorkerProcessPingResponse                                EMsg = 9001
	EMsg_WorkerProcessShutdown                                    EMsg = 9002
	EMsg_DRMWorkerProcess                                         EMsg = 9100
	EMsg_DRMWorkerProcessDRMAndSign                               EMsg = 9100
	EMsg_DRMWorkerProcessDRMAndSignResponse                       EMsg = 9101
	EMsg_DRMWorkerProcessSteamworksInfoRequest                    EMsg = 9102
	EMsg_DRMWorkerProcessSteamworksInfoResponse                   EMsg = 9103
	EMsg_DRMWorkerProcessInstallDRMDLLRequest                     EMsg = 9104
	EMsg_DRMWorkerProcessInstallDRMDLLResponse                    EMsg = 9105
	EMsg_DRMWorkerProcessSecretIdStringRequest                    EMsg = 9106
	EMsg_DRMWorkerProcessSecretIdStringResponse                   EMsg = 9107
	EMsg_DRMWorkerProcessInstallProcessedFilesRequest             EMsg = 9110
	EMsg_DRMWorkerProcessInstallProcessedFilesResponse            EMsg = 9111
	EMsg_DRMWorkerProcessExamineBlobRequest                       EMsg = 9112
	EMsg_DRMWorkerProcessExamineBlobResponse                      EMsg = 9113
	EMsg_DRMWorkerProcessDescribeSecretRequest                    EMsg = 9114
	EMsg_DRMWorkerProcessDescribeSecretResponse                   EMsg = 9115
	EMsg_DRMWorkerProcessBackfillOriginalRequest                  EMsg = 9116
	EMsg_DRMWorkerProcessBackfillOriginalResponse                 EMsg = 9117
	EMsg_DRMWorkerProcessValidateDRMDLLRequest                    EMsg = 9118
	EMsg_DRMWorkerProcessValidateDRMDLLResponse                   EMsg = 9119
	EMsg_DRMWorkerProcessValidateFileRequest                      EMsg = 9120
	EMsg_DRMWorkerProcessValidateFileResponse                     EMsg = 9121
	EMsg_DRMWorkerProcessSplitAndInstallRequest                   EMsg = 9122
	EMsg_DRMWorkerProcessSplitAndInstallResponse                  EMsg = 9123
	EMsg_DRMWorkerProcessGetBlobRequest                           EMsg = 9124
	EMsg_DRMWorkerProcessGetBlobResponse                          EMsg = 9125
	EMsg_DRMWorkerProcessEvaluateCrashRequest                     EMsg = 9126
	EMsg_DRMWorkerProcessEvaluateCrashResponse                    EMsg = 9127
	EMsg_DRMWorkerProcessAnalyzeFileRequest                       EMsg = 9128
	EMsg_DRMWorkerProcessAnalyzeFileResponse                      EMsg = 9129
	EMsg_DRMWorkerProcessUnpackBlobRequest                        EMsg = 9130
	EMsg_DRMWorkerProcessUnpackBlobResponse                       EMsg = 9131
	EMsg_DRMWorkerProcessInstallAllRequest                        EMsg = 9132
	EMsg_DRMWorkerProcessInstallAllResponse                       EMsg = 9133
	EMsg_TestWorkerProcess                                        EMsg = 9200
	EMsg_TestWorkerProcessLoadUnloadModuleRequest                 EMsg = 9200
	EMsg_TestWorkerProcessLoadUnloadModuleResponse                EMsg = 9201
	EMsg_TestWorkerProcessServiceModuleCallRequest                EMsg = 9202
	EMsg_TestWorkerProcessServiceModuleCallResponse               EMsg = 9203
	EMsg_QuestServerBase                                          EMsg = 9300
	EMsg_ClientGetEmoticonList                                    EMsg = 9330
	EMsg_ClientEmoticonList                                       EMsg = 9331
	EMsg_SLCBase                                                  EMsg = 9400
	EMsg_SLCUserSessionStatus                                     EMsg = 9400
	EMsg_SLCRequestUserSessionStatus                              EMsg = 9401
	EMsg_SLCSharedLicensesLockStatus                              EMsg = 9402
	EMsg_ClientSharedLibraryLockStatus                            EMsg = 9405
	EMsg_ClientSharedLibraryStopPlaying                           EMsg = 9406
	EMsg_SLCOwnerLibraryChanged                                   EMsg = 9407
	EMsg_SLCSharedLibraryChanged                                  EMsg = 9408
	EMsg_RemoteClientBase                                         EMsg = 9500
	EMsg_RemoteClientAuth                                         EMsg = 9500 // Deprecated
	EMsg_RemoteClientAuthResponse                                 EMsg = 9501 // Deprecated
	EMsg_RemoteClientAppStatus                                    EMsg = 9502
	EMsg_RemoteClientStartStream                                  EMsg = 9503
	EMsg_RemoteClientStartStreamResponse                          EMsg = 9504
	EMsg_RemoteClientPing                                         EMsg = 9505
	EMsg_RemoteClientPingResponse                                 EMsg = 9506
	EMsg_ClientUnlockStreaming                                    EMsg = 9507
	EMsg_ClientUnlockStreamingResponse                            EMsg = 9508
	EMsg_RemoteClientAcceptEULA                                   EMsg = 9509
	EMsg_RemoteClientGetControllerConfig                          EMsg = 9510
	EMsg_RemoteClientGetControllerConfigResposne                  EMsg = 9511 // Deprecated: renamed to RemoteClientGetControllerConfigResponse
	EMsg_RemoteClientGetControllerConfigResponse                  EMsg = 9511
	EMsg_RemoteClientStreamingEnabled                             EMsg = 9512
	EMsg_ClientUnlockHEVC                                         EMsg = 9513
	EMsg_ClientUnlockHEVCResponse                                 EMsg = 9514
	EMsg_RemoteClientStatusRequest                                EMsg = 9515
	EMsg_RemoteClientStatusResponse                               EMsg = 9516
	EMsg_ClientConcurrentSessionsBase                             EMsg = 9600
	EMsg_ClientPlayingSessionState                                EMsg = 9600
	EMsg_ClientKickPlayingSession                                 EMsg = 9601
	EMsg_ClientBroadcastBase                                      EMsg = 9700
	EMsg_ClientBroadcastInit                                      EMsg = 9700
	EMsg_ClientBroadcastFrames                                    EMsg = 9701
	EMsg_ClientBroadcastDisconnect                                EMsg = 9702
	EMsg_ClientBroadcastScreenshot                                EMsg = 9703
	EMsg_ClientBroadcastUploadConfig                              EMsg = 9704
	EMsg_BaseClient3                                              EMsg = 9800
	EMsg_ClientVoiceCallPreAuthorize                              EMsg = 9800
	EMsg_ClientVoiceCallPreAuthorizeResponse                      EMsg = 9801
	EMsg_ClientServerTimestampRequest                             EMsg = 9802
	EMsg_ClientServerTimestampResponse                            EMsg = 9803
	EMsg_ClientLANP2PBase                                         EMsg = 9900
	EMsg_ClientLANP2PRequestChunk                                 EMsg = 9900
	EMsg_ClientLANP2PRequestChunkResponse                         EMsg = 9901
	EMsg_ClientLANP2PMax                                          EMsg = 9999
	EMsg_BaseWatchdogServer                                       EMsg = 10000
	EMsg_NotifyWatchdog                                           EMsg = 10000
	EMsg_ClientSiteLicenseBase                                    EMsg = 10100
	EMsg_ClientSiteLicenseSiteInfoNotification                    EMsg = 10100
	EMsg_ClientSiteLicenseCheckout                                EMsg = 10101
	EMsg_ClientSiteLicenseCheckoutResponse                        EMsg = 10102
	EMsg_ClientSiteLicenseGetAvailableSeats                       EMsg = 10103
	EMsg_ClientSiteLicenseGetAvailableSeatsResponse               EMsg = 10104
	EMsg_ClientSiteLicenseGetContentCacheInfo                     EMsg = 10105
	EMsg_ClientSiteLicenseGetContentCacheInfoResponse             EMsg = 10106
	EMsg_BaseChatServer                                           EMsg = 12000
	EMsg_ChatServerGetPendingNotificationCount                    EMsg = 12000
	EMsg_ChatServerGetPendingNotificationCountResponse            EMsg = 12001
	EMsg_BaseSecretServer                                         EMsg = 12100
	EMsg_ServerSecretChanged                                      EMsg = 12100
)

var EMsg_name = map[EMsg]string{
	0:     "EMsg_Invalid",
	1:     "EMsg_Multi",
	2:     "EMsg_ProtobufWrapped",
	100:   "EMsg_BaseGeneral",
	113:   "EMsg_DestJobFailed",
	115:   "EMsg_Alert",
	120:   "EMsg_SCIDRequest",
	121:   "EMsg_SCIDResponse",
	123:   "EMsg_JobHeartbeat",
	124:   "EMsg_HubConnect",
	126:   "EMsg_Subscribe",
	127:   "EMsg_RouteMessage",
	128:   "EMsg_RemoteSysID",
	129:   "EMsg_AMCreateAccountResponse",
	130:   "EMsg_WGRequest",
	131:   "EMsg_WGResponse",
	132:   "EMsg_KeepAlive",
	133:   "EMsg_WebAPIJobRequest",
	134:   "EMsg_WebAPIJobResponse",
	135:   "EMsg_ClientSessionStart",
	136:   "EMsg_ClientSessionEnd",
	137:   "EMsg_ClientSessionUpdateAuthTicket",
	138:   "EMsg_StatsDeprecated",
	139:   "EMsg_Ping",
	140:   "EMsg_PingResponse",
	141:   "EMsg_Stats",
	142:   "EMsg_RequestFullStatsBlock",
	143:   "EMsg_LoadDBOCacheItem",
	144:   "EMsg_LoadDBOCacheItemResponse",
	145:   "EMsg_InvalidateDBOCacheItems",
	146:   "EMsg_ServiceMethod",
	147:   "EMsg_ServiceMethodResponse",
	148:   "EMsg_ClientPackageVersions",
	149:   "EMsg_TimestampRequest",
	150:   "EMsg_TimestampResponse",
	151:   "EMsg_ServiceMethodCallFromClient",
	152:   "EMsg_ServiceMethodSendToClient",
	200:   "EMsg_BaseShell",
	201:   "EMsg_Exit",
	202:   "EMsg_DirRequest",
	203:   "EMsg_DirResponse",
	204:   "EMsg_ZipRequest",
	205:   "EMsg_ZipResponse",
	215:   "EMsg_UpdateRecordResponse",
	221:   "EMsg_UpdateCreditCardRequest",
	225:   "EMsg_UpdateUserBanResponse",
	226:   "EMsg_PrepareToExit",
	227:   "EMsg_ContentDescriptionUpdate",
	228:   "EMsg_TestResetServer",
	229:   "EMsg_UniverseChanged",
	230:   "EMsg_ShellConfigInfoUpdate",
	233:   "EMsg_RequestWindowsEventLogEntries",
	234:   "EMsg_ProvideWindowsEventLogEntries",
	235:   "EMsg_ShellSearchLogs",
	236:   "EMsg_ShellSearchLogsResponse",
	237:   "EMsg_ShellCheckWindowsUpdates",
	238:   "EMsg_ShellCheckWindowsUpdatesResponse",
	239:   "EMsg_ShellFlushUserLicenseCache",
	240:   "EMsg_TestFlushDelayedSQL",
	241:   "EMsg_TestFlushDelayedSQLResponse",
	242:   "EMsg_EnsureExecuteScheduledTask_TEST",
	243:   "EMsg_EnsureExecuteScheduledTaskResponse_TEST",
	244:   "EMsg_UpdateScheduledTaskEnableState_TEST",
	245:   "EMsg_UpdateScheduledTaskEnableStateResponse_TEST",
	246:   "EMsg_ContentDescriptionDeltaUpdate",
	300:   "EMsg_BaseGM",
	301:   "EMsg_ShellFailed",
	307:   "EMsg_ExitShells",
	308:   "EMsg_ExitShell",
	309:   "EMsg_GracefulExitShell",
	314:   "EMsg_NotifyWatchdog",
	316:   "EMsg_LicenseProcessingComplete",
	317:   "EMsg_SetTestFlag",
	318:   "EMsg_QueuedEmailsComplete",
	319:   "EMsg_GMReportPHPError",
	320:   "EMsg_GMDRMSync",
	321:   "EMsg_PhysicalBoxInventory",
	322:   "EMsg_UpdateConfigFile",
	323:   "EMsg_TestInitDB",
	324:   "EMsg_GMWriteConfigToSQL",
	325:   "EMsg_GMLoadActivationCodes",
	326:   "EMsg_GMQueueForFBS",
	327:   "EMsg_GMSchemaConversionResults",
	328:   "EMsg_GMSchemaConversionResultsResponse",
	329:   "EMsg_GMWriteShellFailureToSQL",
	330:   "EMsg_GMWriteStatsToSOS",
	331:   "EMsg_GMGetServiceMethodRouting",
	332:   "EMsg_GMGetServiceMethodRoutingResponse",
	333:   "EMsg_GMConvertUserWallets",
	334:   "EMsg_GMTestNextBuildSchemaConversion",
	335:   "EMsg_GMTestNextBuildSchemaConversionResponse",
	336:   "EMsg_ExpectShellRestart",
	337:   "EMsg_HotFixProgress",
	400:   "EMsg_BaseAIS",
	401:   "EMsg_AISRefreshContentDescription",
	402:   "EMsg_AISRequestContentDescription",
	403:   "EMsg_AISUpdateAppInfo",
	404:   "EMsg_AISUpdatePackageInfo",
	405:   "EMsg_AISGetPackageChangeNumber",
	406:   "EMsg_AISGetPackageChangeNumberResponse",
	407:   "EMsg_AISAppInfoTableChanged",
	408:   "EMsg_AISUpdatePackageCostsResponse",
	409:   "EMsg_AISCreateMarketingMessage",
	410:   "EMsg_AISCreateMarketingMessageResponse",
	411:   "EMsg_AISGetMarketingMessage",
	412:   "EMsg_AISGetMarketingMessageResponse",
	413:   "EMsg_AISUpdateMarketingMessage",
	414:   "EMsg_AISUpdateMarketingMessageResponse",
	415:   "EMsg_AISRequestMarketingMessageUpdate",
	416:   "EMsg_AISDeleteMarketingMessage",
	419:   "EMsg_AISGetMarketingTreatments",
	420:   "EMsg_AISGetMarketingTreatmentsResponse",
	421:   "EMsg_AISRequestMarketingTreatmentUpdate",
	422:   "EMsg_AISTestAddPackage",
	423:   "EMsg_AIGetAppGCFlags",
	424:   "EMsg_AIGetAppGCFlagsResponse",
	425:   "EMsg_AIGetAppList",
	426:   "EMsg_AIGetAppListResponse",
	427:   "EMsg_AIGetAppInfo",
	428:   "EMsg_AIGetAppInfoResponse",
	429:   "EMsg_AISGetCouponDefinition",
	430:   "EMsg_AISGetCouponDefinitionResponse",
	431:   "EMsg_AISUpdateSlaveContentDescription",
	432:   "EMsg_AISUpdateSlaveContentDescriptionResponse",
	433:   "EMsg_AISTestEnableGC",
	500:   "EMsg_BaseAM",
	504:   "EMsg_AMUpdateUserBanRequest",
	505:   "EMsg_AMAddLicense",
	507:   "EMsg_AMBeginProcessingLicenses",
	508:   "EMsg_AMSendSystemIMToUser",
	509:   "EMsg_AMExtendLicense",
	510:   "EMsg_AMAddMinutesToLicense",
	511:   "EMsg_AMCancelLicense",
	512:   "EMsg_AMInitPurchase",
	513:   "EMsg_AMPurchaseResponse",
	514:   "EMsg_AMGetFinalPrice",
	515:   "EMsg_AMGetFinalPriceResponse",
	516:   "EMsg_AMGetLegacyGameKey",
	517:   "EMsg_AMGetLegacyGameKeyResponse",
	518:   "EMsg_AMFindHungTransactions",
	519:   "EMsg_AMSetAccountTrustedRequest",
	521:   "EMsg_AMCompletePurchase",
	522:   "EMsg_AMCancelPurchase",
	523:   "EMsg_AMNewChallenge",
	524:   "EMsg_AMLoadOEMTickets",
	525:   "EMsg_AMFixPendingPurchase",
	526:   "EMsg_AMFixPendingPurchaseResponse",
	527:   "EMsg_AMIsUserBanned",
	528:   "EMsg_AMRegisterKey",
	529:   "EMsg_AMLoadActivationCodes",
	530:   "EMsg_AMLoadActivationCodesResponse",
	531:   "EMsg_AMLookupKeyResponse",
	532:   "EMsg_AMLookupKey",
	533:   "EMsg_AMChatCleanup",
	534:   "EMsg_AMClanCleanup",
	535:   "EMsg_AMFixPendingRefund",
	536:   "EMsg_AMReverseChargeback",
	537:   "EMsg_AMReverseChargebackResponse",
	538:   "EMsg_AMClanCleanupList",
	539:   "EMsg_AMGetLicenses",
	540:   "EMsg_AMGetLicensesResponse",
	541:   "EMsg_AMSendCartRepurchase",
	542:   "EMsg_AMSendCartRepurchaseResponse",
	550:   "EMsg_AllowUserToPlayQuery",
	551:   "EMsg_AllowUserToPlayResponse",
	552:   "EMsg_AMVerfiyUser",
	553:   "EMsg_AMClientNotPlaying",
	554:   "EMsg_ClientRequestFriendship",
	555:   "EMsg_AMRelayPublishStatus",
	556:   "EMsg_AMResetCommunityContent",
	557:   "EMsg_AMPrimePersonaStateCache",
	558:   "EMsg_AMAllowUserContentQuery",
	559:   "EMsg_AMAllowUserContentResponse",
	560:   "EMsg_AMInitPurchaseResponse",
	561:   "EMsg_AMRevokePurchaseResponse",
	562:   "EMsg_AMLockProfile",
	563:   "EMsg_AMRefreshGuestPasses",
	564:   "EMsg_AMInviteUserToClan",
	565:   "EMsg_AMAcknowledgeClanInvite",
	566:   "EMsg_AMGrantGuestPasses",
	567:   "EMsg_AMClanDataUpdated",
	568:   "EMsg_AMReloadAccount",
	569:   "EMsg_AMClientChatMsgRelay",
	570:   "EMsg_AMChatMulti",
	571:   "EMsg_AMClientChatInviteRelay",
	572:   "EMsg_AMChatInvite",
	573:   "EMsg_AMClientJoinChatRelay",
	574:   "EMsg_AMClientChatMemberInfoRelay",
	575:   "EMsg_AMPublishChatMemberInfo",
	576:   "EMsg_AMClientAcceptFriendInvite",
	577:   "EMsg_AMChatEnter",
	578:   "EMsg_AMClientPublishRemovalFromSource",
	579:   "EMsg_AMChatActionResult",
	580:   "EMsg_AMFindAccounts",
	581:   "EMsg_AMFindAccountsResponse",
	582:   "EMsg_AMRequestAccountData",
	583:   "EMsg_AMRequestAccountDataResponse",
	584:   "EMsg_AMSetAccountFlags",
	586:   "EMsg_AMCreateClan",
	587:   "EMsg_AMCreateClanResponse",
	588:   "EMsg_AMGetClanDetails",
	589:   "EMsg_AMGetClanDetailsResponse",
	590:   "EMsg_AMSetPersonaName",
	591:   "EMsg_AMSetAvatar",
	592:   "EMsg_AMAuthenticateUser",
	593:   "EMsg_AMAuthenticateUserResponse",
	594:   "EMsg_AMGetAccountFriendsCount",
	595:   "EMsg_AMGetAccountFriendsCountResponse",
	596:   "EMsg_AMP2PIntroducerMessage",
	597:   "EMsg_ClientChatAction",
	598:   "EMsg_AMClientChatActionRelay",
	600:   "EMsg_BaseVS",
	601:   "EMsg_VACResponse",
	602:   "EMsg_ReqChallengeTest",
	604:   "EMsg_VSMarkCheat",
	605:   "EMsg_VSAddCheat",
	606:   "EMsg_VSPurgeCodeModDB",
	607:   "EMsg_VSGetChallengeResults",
	608:   "EMsg_VSChallengeResultText",
	609:   "EMsg_VSReportLingerer",
	610:   "EMsg_VSRequestManagedChallenge",
	611:   "EMsg_VSLoadDBFinished",
	625:   "EMsg_BaseDRMS",
	628:   "EMsg_DRMBuildBlobRequest",
	629:   "EMsg_DRMBuildBlobResponse",
	630:   "EMsg_DRMResolveGuidRequest",
	631:   "EMsg_DRMResolveGuidResponse",
	633:   "EMsg_DRMVariabilityReport",
	634:   "EMsg_DRMVariabilityReportResponse",
	635:   "EMsg_DRMStabilityReport",
	636:   "EMsg_DRMStabilityReportResponse",
	637:   "EMsg_DRMDetailsReportRequest",
	638:   "EMsg_DRMDetailsReportResponse",
	639:   "EMsg_DRMProcessFile",
	640:   "EMsg_DRMAdminUpdate",
	641:   "EMsg_DRMAdminUpdateResponse",
	642:   "EMsg_DRMSync",
	643:   "EMsg_DRMSyncResponse",
	644:   "EMsg_DRMProcessFileResponse",
	645:   "EMsg_DRMEmptyGuidCache",
	646:   "EMsg_DRMEmptyGuidCacheResponse",
	650:   "EMsg_BaseCS",
	652:   "EMsg_CSUserContentRequest",
	700:   "EMsg_BaseClient",
	701:   "EMsg_ClientLogOn_Deprecated",
	702:   "EMsg_ClientAnonLogOn_Deprecated",
	703:   "EMsg_ClientHeartBeat",
	704:   "EMsg_ClientVACResponse",
	705:   "EMsg_ClientGamesPlayed_obsolete",
	706:   "EMsg_ClientLogOff",
	707:   "EMsg_ClientNoUDPConnectivity",
	708:   "EMsg_ClientInformOfCreateAccount",
	709:   "EMsg_ClientAckVACBan",
	710:   "EMsg_ClientConnectionStats",
	711:   "EMsg_ClientInitPurchase",
	712:   "EMsg_ClientPingResponse",
	714:   "EMsg_ClientRemoveFriend",
	715:   "EMsg_ClientGamesPlayedNoDataBlob",
	716:   "EMsg_ClientChangeStatus",
	717:   "EMsg_ClientVacStatusResponse",
	718:   "EMsg_ClientFriendMsg",
	719:   "EMsg_ClientGameConnect_obsolete",
	720:   "EMsg_ClientGamesPlayed2_obsolete",
	721:   "EMsg_ClientGameEnded_obsolete",
	722:   "EMsg_ClientGetFinalPrice",
	726:   "EMsg_ClientSystemIM",
	727:   "EMsg_ClientSystemIMAck",
	728:   "EMsg_ClientGetLicenses",
	729:   "EMsg_ClientCancelLicense",
	730:   "EMsg_ClientGetLegacyGameKey",
	731:   "EMsg_ClientContentServerLogOn_Deprecated",
	732:   "EMsg_ClientAckVACBan2",
	735:   "EMsg_ClientAckMessageByGID",
	736:   "EMsg_ClientGetPurchaseReceipts",
	737:   "EMsg_ClientAckPurchaseReceipt",
	738:   "EMsg_ClientGamesPlayed3_obsolete",
	739:   "EMsg_ClientSendGuestPass",
	740:   "EMsg_ClientAckGuestPass",
	741:   "EMsg_ClientRedeemGuestPass",
	742:   "EMsg_ClientGamesPlayed",
	743:   "EMsg_ClientRegisterKey",
	744:   "EMsg_ClientInviteUserToClan",
	745:   "EMsg_ClientAcknowledgeClanInvite",
	746:   "EMsg_ClientPurchaseWithMachineID",
	747:   "EMsg_ClientAppUsageEvent",
	748:   "EMsg_ClientGetGiftTargetList",
	749:   "EMsg_ClientGetGiftTargetListResponse",
	751:   "EMsg_ClientLogOnResponse",
	753:   "EMsg_ClientVACChallenge",
	755:   "EMsg_ClientSetHeartbeatRate",
	756:   "EMsg_ClientNotLoggedOnDeprecated",
	757:   "EMsg_ClientLoggedOff",
	758:   "EMsg_GSApprove",
	759:   "EMsg_GSDeny",
	760:   "EMsg_GSKick",
	761:   "EMsg_ClientCreateAcctResponse",
	763:   "EMsg_ClientPurchaseResponse",
	764:   "EMsg_ClientPing",
	765:   "EMsg_ClientNOP",
	766:   "EMsg_ClientPersonaState",
	767:   "EMsg_ClientFriendsList",
	768:   "EMsg_ClientAccountInfo",
	770:   "EMsg_ClientVacStatusQuery",
	771:   "EMsg_ClientNewsUpdate",
	773:   "EMsg_ClientGameConnectDeny",
	774:   "EMsg_GSStatusReply",
	775:   "EMsg_ClientGetFinalPriceResponse",
	779:   "EMsg_ClientGameConnectTokens",
	780:   "EMsg_ClientLicenseList",
	781:   "EMsg_ClientCancelLicenseResponse",
	782:   "EMsg_ClientVACBanStatus",
	783:   "EMsg_ClientCMList",
	784:   "EMsg_ClientEncryptPct",
	785:   "EMsg_ClientGetLegacyGameKeyResponse",
	786:   "EMsg_ClientFavoritesList",
	787:   "EMsg_CSUserContentApprove",
	788:   "EMsg_CSUserContentDeny",
	789:   "EMsg_ClientInitPurchaseResponse",
	791:   "EMsg_ClientAddFriend",
	792:   "EMsg_ClientAddFriendResponse",
	793:   "EMsg_ClientInviteFriend",
	794:   "EMsg_ClientInviteFriendResponse",
	795:   "EMsg_ClientSendGuestPassResponse",
	796:   "EMsg_ClientAckGuestPassResponse",
	797:   "EMsg_ClientRedeemGuestPassResponse",
	798:   "EMsg_ClientUpdateGuestPassesList",
	799:   "EMsg_ClientChatMsg",
	800:   "EMsg_ClientChatInvite",
	801:   "EMsg_ClientJoinChat",
	802:   "EMsg_ClientChatMemberInfo",
	803:   "EMsg_ClientLogOnWithCredentials_Deprecated",
	805:   "EMsg_ClientPasswordChangeResponse",
	807:   "EMsg_ClientChatEnter",
	808:   "EMsg_ClientFriendRemovedFromSource",
	809:   "EMsg_ClientCreateChat",
	810:   "EMsg_ClientCreateChatResponse",
	811:   "EMsg_ClientUpdateChatMetadata",
	813:   "EMsg_ClientP2PIntroducerMessage",
	814:   "EMsg_ClientChatActionResult",
	815:   "EMsg_ClientRequestFriendData",
	818:   "EMsg_ClientGetUserStats",
	819:   "EMsg_ClientGetUserStatsResponse",
	820:   "EMsg_ClientStoreUserStats",
	821:   "EMsg_ClientStoreUserStatsResponse",
	822:   "EMsg_ClientClanState",
	830:   "EMsg_ClientServiceModule",
	831:   "EMsg_ClientServiceCall",
	832:   "EMsg_ClientServiceCallResponse",
	833:   "EMsg_ClientPackageInfoRequest",
	834:   "EMsg_ClientPackageInfoResponse",
	839:   "EMsg_ClientNatTraversalStatEvent",
	840:   "EMsg_ClientAppInfoRequest",
	841:   "EMsg_ClientAppInfoResponse",
	842:   "EMsg_ClientSteamUsageEvent",
	845:   "EMsg_ClientCheckPassword",
	846:   "EMsg_ClientResetPassword",
	848:   "EMsg_ClientCheckPasswordResponse",
	849:   "EMsg_ClientResetPasswordResponse",
	850:   "EMsg_ClientSessionToken",
	851:   "EMsg_ClientDRMProblemReport",
	855:   "EMsg_ClientSetIgnoreFriend",
	856:   "EMsg_ClientSetIgnoreFriendResponse",
	857:   "EMsg_ClientGetAppOwnershipTicket",
	858:   "EMsg_ClientGetAppOwnershipTicketResponse",
	860:   "EMsg_ClientGetLobbyListResponse",
	861:   "EMsg_ClientGetLobbyMetadata",
	862:   "EMsg_ClientGetLobbyMetadataResponse",
	863:   "EMsg_ClientVTTCert",
	866:   "EMsg_ClientAppInfoUpdate",
	867:   "EMsg_ClientAppInfoChanges",
	880:   "EMsg_ClientServerList",
	891:   "EMsg_ClientEmailChangeResponse",
	892:   "EMsg_ClientSecretQAChangeResponse",
	896:   "EMsg_ClientDRMBlobRequest",
	897:   "EMsg_ClientDRMBlobResponse",
	898:   "EMsg_ClientLookupKey",
	899:   "EMsg_ClientLookupKeyResponse",
	900:   "EMsg_BaseGameServer",
	901:   "EMsg_GSDisconnectNotice",
	903:   "EMsg_GSStatus",
	905:   "EMsg_GSUserPlaying",
	906:   "EMsg_GSStatus2",
	907:   "EMsg_GSStatusUpdate_Unused",
	908:   "EMsg_GSServerType",
	909:   "EMsg_GSPlayerList",
	910:   "EMsg_GSGetUserAchievementStatus",
	911:   "EMsg_GSGetUserAchievementStatusResponse",
	918:   "EMsg_GSGetPlayStats",
	919:   "EMsg_GSGetPlayStatsResponse",
	920:   "EMsg_GSGetUserGroupStatus",
	921:   "EMsg_AMGetUserGroupStatus",
	922:   "EMsg_AMGetUserGroupStatusResponse",
	923:   "EMsg_GSGetUserGroupStatusResponse",
	936:   "EMsg_GSGetReputation",
	937:   "EMsg_GSGetReputationResponse",
	938:   "EMsg_GSAssociateWithClan",
	939:   "EMsg_GSAssociateWithClanResponse",
	940:   "EMsg_GSComputeNewPlayerCompatibility",
	941:   "EMsg_GSComputeNewPlayerCompatibilityResponse",
	1000:  "EMsg_BaseAdmin",
	1004:  "EMsg_AdminCmdResponse",
	1005:  "EMsg_AdminLogListenRequest",
	1006:  "EMsg_AdminLogEvent",
	1007:  "EMsg_LogSearchRequest",
	1008:  "EMsg_LogSearchResponse",
	1009:  "EMsg_LogSearchCancel",
	1010:  "EMsg_UniverseData",
	1014:  "EMsg_RequestStatHistory",
	1015:  "EMsg_StatHistory",
	1017:  "EMsg_AdminPwLogon",
	1018:  "EMsg_AdminPwLogonResponse",
	1019:  "EMsg_AdminSpew",
	1020:  "EMsg_AdminConsoleTitle",
	1023:  "EMsg_AdminGCSpew",
	1024:  "EMsg_AdminGCCommand",
	1025:  "EMsg_AdminGCGetCommandList",
	1026:  "EMsg_AdminGCGetCommandListResponse",
	1027:  "EMsg_FBSConnectionData",
	1028:  "EMsg_AdminMsgSpew",
	1100:  "EMsg_BaseFBS",
	1101:  "EMsg_FBSVersionInfo",
	1102:  "EMsg_FBSForceRefresh",
	1103:  "EMsg_FBSForceBounce",
	1104:  "EMsg_FBSDeployPackage",
	1105:  "EMsg_FBSDeployResponse",
	1106:  "EMsg_FBSUpdateBootstrapper",
	1107:  "EMsg_FBSSetState",
	1108:  "EMsg_FBSApplyOSUpdates",
	1109:  "EMsg_FBSRunCMDScript",
	1110:  "EMsg_FBSRebootBox",
	1111:  "EMsg_FBSSetBigBrotherMode",
	1112:  "EMsg_FBSMinidumpServer",
	1113:  "EMsg_FBSSetShellCount_obsolete",
	1114:  "EMsg_FBSDeployHotFixPackage",
	1115:  "EMsg_FBSDeployHotFixResponse",
	1116:  "EMsg_FBSDownloadHotFix",
	1117:  "EMsg_FBSDownloadHotFixResponse",
	1118:  "EMsg_FBSUpdateTargetConfigFile",
	1119:  "EMsg_FBSApplyAccountCred",
	1120:  "EMsg_FBSApplyAccountCredResponse",
	1121:  "EMsg_FBSSetShellCount",
	1122:  "EMsg_FBSTerminateShell",
	1123:  "EMsg_FBSQueryGMForRequest",
	1124:  "EMsg_FBSQueryGMResponse",
	1125:  "EMsg_FBSTerminateZombies",
	1126:  "EMsg_FBSInfoFromBootstrapper",
	1127:  "EMsg_FBSRebootBoxResponse",
	1128:  "EMsg_FBSBootstrapperPackageRequest",
	1129:  "EMsg_FBSBootstrapperPackageResponse",
	1130:  "EMsg_FBSBootstrapperGetPackageChunk",
	1131:  "EMsg_FBSBootstrapperGetPackageChunkResponse",
	1132:  "EMsg_FBSBootstrapperPackageTransferProgress",
	1133:  "EMsg_FBSRestartBootstrapper",
	1134:  "EMsg_FBSPauseFrozenDumps",
	1200:  "EMsg_BaseFileXfer",
	1201:  "EMsg_FileXferResponse",
	1202:  "EMsg_FileXferData",
	1203:  "EMsg_FileXferEnd",
	1204:  "EMsg_FileXferDataAck",
	1300:  "EMsg_BaseChannelAuth",
	1301:  "EMsg_ChannelAuthResponse",
	1302:  "EMsg_ChannelAuthResult",
	1303:  "EMsg_ChannelEncryptRequest",
	1304:  "EMsg_ChannelEncryptResponse",
	1305:  "EMsg_ChannelEncryptResult",
	1400:  "EMsg_BaseBS",
	1401:  "EMsg_BSPurchaseStart",
	1402:  "EMsg_BSPurchaseResponse",
	1403:  "EMsg_BSAuthenticateCCTrans",
	1404:  "EMsg_BSAuthenticateCCTransResponse",
	1406:  "EMsg_BSSettleComplete",
	1407:  "EMsg_BSBannedRequest",
	1408:  "EMsg_BSInitPayPalTxn",
	1409:  "EMsg_BSInitPayPalTxnResponse",
	1410:  "EMsg_BSGetPayPalUserInfo",
	1411:  "EMsg_BSGetPayPalUserInfoResponse",
	1413:  "EMsg_BSRefundTxn",
	1414:  "EMsg_BSRefundTxnResponse",
	1415:  "EMsg_BSGetEvents",
	1416:  "EMsg_BSChaseRFRRequest",
	1417:  "EMsg_BSPaymentInstrBan",
	1418:  "EMsg_BSPaymentInstrBanResponse",
	1419:  "EMsg_BSProcessGCReports",
	1420:  "EMsg_BSProcessPPReports",
	1421:  "EMsg_BSInitGCBankXferTxn",
	1422:  "EMsg_BSInitGCBankXferTxnResponse",
	1423:  "EMsg_BSQueryGCBankXferTxn",
	1424:  "EMsg_BSQueryGCBankXferTxnResponse",
	1425:  "EMsg_BSCommitGCTxn",
	1426:  "EMsg_BSQueryTransactionStatus",
	1427:  "EMsg_BSQueryTransactionStatusResponse",
	1428:  "EMsg_BSQueryCBOrderStatus",
	1429:  "EMsg_BSQueryCBOrderStatusResponse",
	1430:  "EMsg_BSRunRedFlagReport",
	1431:  "EMsg_BSQueryPaymentInstUsage",
	1432:  "EMsg_BSQueryPaymentInstResponse",
	1433:  "EMsg_BSQueryTxnExtendedInfo",
	1434:  "EMsg_BSQueryTxnExtendedInfoResponse",
	1435:  "EMsg_BSUpdateConversionRates",
	1436:  "EMsg_BSProcessUSBankReports",
	1437:  "EMsg_BSPurchaseRunFraudChecks",
	1438:  "EMsg_BSPurchaseRunFraudChecksResponse",
	1439:  "EMsg_BSStartShippingJobs",
	1440:  "EMsg_BSQueryBankInformation",
	1441:  "EMsg_BSQueryBankInformationResponse",
	1445:  "EMsg_BSValidateXsollaSignature",
	1446:  "EMsg_BSValidateXsollaSignatureResponse",
	1448:  "EMsg_BSQiwiWalletInvoice",
	1449:  "EMsg_BSQiwiWalletInvoiceResponse",
	1450:  "EMsg_BSUpdateInventoryFromProPack",
	1451:  "EMsg_BSUpdateInventoryFromProPackResponse",
	1452:  "EMsg_BSSendShippingRequest",
	1453:  "EMsg_BSSendShippingRequestResponse",
	1454:  "EMsg_BSGetProPackOrderStatus",
	1455:  "EMsg_BSGetProPackOrderStatusResponse",
	1456:  "EMsg_BSCheckJobRunning",
	1457:  "EMsg_BSCheckJobRunningResponse",
	1458:  "EMsg_BSResetPackagePurchaseRateLimit",
	1459:  "EMsg_BSResetPackagePurchaseRateLimitResponse",
	1460:  "EMsg_BSUpdatePaymentData",
	1461:  "EMsg_BSUpdatePaymentDataResponse",
	1462:  "EMsg_BSGetBillingAddress",
	1463:  "EMsg_BSGetBillingAddressResponse",
	1464:  "EMsg_BSGetCreditCardInfo",
	1465:  "EMsg_BSGetCreditCardInfoResponse",
	1468:  "EMsg_BSRemoveExpiredPaymentData",
	1469:  "EMsg_BSRemoveExpiredPaymentDataResponse",
	1470:  "EMsg_BSConvertToCurrentKeys",
	1471:  "EMsg_BSConvertToCurrentKeysResponse",
	1472:  "EMsg_BSInitPurchase",
	1473:  "EMsg_BSInitPurchaseResponse",
	1474:  "EMsg_BSCompletePurchase",
	1475:  "EMsg_BSCompletePurchaseResponse",
	1476:  "EMsg_BSPruneCardUsageStats",
	1477:  "EMsg_BSPruneCardUsageStatsResponse",
	1478:  "EMsg_BSStoreBankInformation",
	1479:  "EMsg_BSStoreBankInformationResponse",
	1480:  "EMsg_BSVerifyPOSAKey",
	1481:  "EMsg_BSVerifyPOSAKeyResponse",
	1482:  "EMsg_BSReverseRedeemPOSAKey",
	1483:  "EMsg_BSReverseRedeemPOSAKeyResponse",
	1484:  "EMsg_BSQueryFindCreditCard",
	1485:  "EMsg_BSQueryFindCreditCardResponse",
	1486:  "EMsg_BSStatusInquiryPOSAKey",
	1487:  "EMsg_BSStatusInquiryPOSAKeyResponse",
	1488:  "EMsg_BSValidateMoPaySignature",
	1489:  "EMsg_BSValidateMoPaySignatureResponse",
	1490:  "EMsg_BSMoPayConfirmProductDelivery",
	1491:  "EMsg_BSMoPayConfirmProductDeliveryResponse",
	1492:  "EMsg_BSGenerateMoPayMD5",
	1493:  "EMsg_BSGenerateMoPayMD5Response",
	1494:  "EMsg_BSBoaCompraConfirmProductDelivery",
	1495:  "EMsg_BSBoaCompraConfirmProductDeliveryResponse",
	1496:  "EMsg_BSGenerateBoaCompraMD5",
	1497:  "EMsg_BSGenerateBoaCompraMD5Response",
	1498:  "EMsg_BSCommitWPTxn",
	1499:  "EMsg_BSCommitAdyenTxn",
	1500:  "EMsg_BaseATS",
	1501:  "EMsg_ATSStartStressTest",
	1502:  "EMsg_ATSStopStressTest",
	1503:  "EMsg_ATSRunFailServerTest",
	1504:  "EMsg_ATSUFSPerfTestTask",
	1505:  "EMsg_ATSUFSPerfTestResponse",
	1506:  "EMsg_ATSCycleTCM",
	1507:  "EMsg_ATSInitDRMSStressTest",
	1508:  "EMsg_ATSCallTest",
	1509:  "EMsg_ATSCallTestReply",
	1510:  "EMsg_ATSStartExternalStress",
	1511:  "EMsg_ATSExternalStressJobStart",
	1512:  "EMsg_ATSExternalStressJobQueued",
	1513:  "EMsg_ATSExternalStressJobRunning",
	1514:  "EMsg_ATSExternalStressJobStopped",
	1515:  "EMsg_ATSExternalStressJobStopAll",
	1516:  "EMsg_ATSExternalStressActionResult",
	1517:  "EMsg_ATSStarted",
	1518:  "EMsg_ATSCSPerfTestTask",
	1519:  "EMsg_ATSCSPerfTestResponse",
	1600:  "EMsg_BaseDP",
	1601:  "EMsg_DPSetPublishingState",
	1602:  "EMsg_DPGamePlayedStats",
	1603:  "EMsg_DPUniquePlayersStat",
	1604:  "EMsg_DPStreamingUniquePlayersStat",
	1605:  "EMsg_DPVacInfractionStats",
	1606:  "EMsg_DPVacBanStats",
	1607:  "EMsg_DPBlockingStats",
	1608:  "EMsg_DPNatTraversalStats",
	1609:  "EMsg_DPSteamUsageEvent",
	1610:  "EMsg_DPVacCertBanStats",
	1611:  "EMsg_DPVacCafeBanStats",
	1612:  "EMsg_DPCloudStats",
	1613:  "EMsg_DPAchievementStats",
	1614:  "EMsg_DPAccountCreationStats",
	1615:  "EMsg_DPGetPlayerCount",
	1616:  "EMsg_DPGetPlayerCountResponse",
	1617:  "EMsg_DPGameServersPlayersStats",
	1618:  "EMsg_DPDownloadRateStatistics",
	1619:  "EMsg_DPFacebookStatistics",
	1620:  "EMsg_ClientDPCheckSpecialSurvey",
	1621:  "EMsg_ClientDPCheckSpecialSurveyResponse",
	1622:  "EMsg_ClientDPSendSpecialSurveyResponse",
	1623:  "EMsg_ClientDPSendSpecialSurveyResponseReply",
	1624:  "EMsg_DPStoreSaleStatistics",
	1625:  "EMsg_ClientDPUpdateAppJobReport",
	1627:  "EMsg_ClientDPSteam2AppStarted",
	1626:  "EMsg_DPUpdateContentEvent",
	1628:  "EMsg_DPPartnerMicroTxns",
	1629:  "EMsg_DPPartnerMicroTxnsResponse",
	1630:  "EMsg_ClientDPContentStatsReport",
	1631:  "EMsg_DPVRUniquePlayersStat",
	1700:  "EMsg_BaseCM",
	1701:  "EMsg_CMSetAllowState",
	1702:  "EMsg_CMSpewAllowState",
	1703:  "EMsg_CMSessionRejected",
	1704:  "EMsg_CMSetSecrets",
	1705:  "EMsg_CMGetSecrets",
	1800:  "EMsg_BaseDSS",
	1801:  "EMsg_DSSNewFile",
	1802:  "EMsg_DSSCurrentFileList",
	1803:  "EMsg_DSSSynchList",
	1804:  "EMsg_DSSSynchListResponse",
	1805:  "EMsg_DSSSynchSubscribe",
	1806:  "EMsg_DSSSynchUnsubscribe",
	1900:  "EMsg_BaseEPM",
	1901:  "EMsg_EPMStartProcess",
	1902:  "EMsg_EPMStopProcess",
	1903:  "EMsg_EPMRestartProcess",
	2200:  "EMsg_BaseGC",
	2201:  "EMsg_AMRelayToGC",
	2202:  "EMsg_GCUpdatePlayedState",
	2203:  "EMsg_GCCmdRevive",
	2204:  "EMsg_GCCmdBounce",
	2205:  "EMsg_GCCmdForceBounce",
	2206:  "EMsg_GCCmdDown",
	2207:  "EMsg_GCCmdDeploy",
	2208:  "EMsg_GCCmdDeployResponse",
	2209:  "EMsg_GCCmdSwitch",
	2210:  "EMsg_AMRefreshSessions",
	2211:  "EMsg_GCUpdateGSState",
	2212:  "EMsg_GCAchievementAwarded",
	2213:  "EMsg_GCSystemMessage",
	2214:  "EMsg_GCValidateSession",
	2215:  "EMsg_GCValidateSessionResponse",
	2216:  "EMsg_GCCmdStatus",
	2217:  "EMsg_GCRegisterWebInterfaces",
	2218:  "EMsg_GCGetAccountDetails",
	2219:  "EMsg_GCInterAppMessage",
	2220:  "EMsg_GCGetEmailTemplate",
	2221:  "EMsg_GCGetEmailTemplateResponse",
	2222:  "EMsg_ISRelayToGCH",
	2223:  "EMsg_GCHRelayClientToIS",
	2224:  "EMsg_GCHUpdateSession",
	2225:  "EMsg_GCHRequestUpdateSession",
	2226:  "EMsg_GCHRequestStatus",
	2227:  "EMsg_GCHRequestStatusResponse",
	2228:  "EMsg_GCHAccountVacStatusChange",
	2229:  "EMsg_GCHSpawnGC",
	2230:  "EMsg_GCHSpawnGCResponse",
	2231:  "EMsg_GCHKillGC",
	2232:  "EMsg_GCHKillGCResponse",
	2233:  "EMsg_GCHAccountTradeBanStatusChange",
	2234:  "EMsg_GCHAccountLockStatusChange",
	2235:  "EMsg_GCHVacVerificationChange",
	2236:  "EMsg_GCHAccountPhoneNumberChange",
	2237:  "EMsg_GCHAccountTwoFactorChange",
	2238:  "EMsg_GCHInviteUserToLobby",
	2500:  "EMsg_BaseP2P",
	2502:  "EMsg_P2PIntroducerMessage",
	2900:  "EMsg_BaseSM",
	2902:  "EMsg_SMExpensiveReport",
	2903:  "EMsg_SMHourlyReport",
	2904:  "EMsg_SMFishingReport",
	2905:  "EMsg_SMPartitionRenames",
	2906:  "EMsg_SMMonitorSpace",
	2907:  "EMsg_SMTestNextBuildSchemaConversion",
	2908:  "EMsg_SMTestNextBuildSchemaConversionResponse",
	3000:  "EMsg_BaseTest",
	3001:  "EMsg_JobHeartbeatTest",
	3002:  "EMsg_JobHeartbeatTestResponse",
	3100:  "EMsg_BaseFTSRange",
	3101:  "EMsg_FTSGetBrowseCounts",
	3102:  "EMsg_FTSGetBrowseCountsResponse",
	3103:  "EMsg_FTSBrowseClans",
	3104:  "EMsg_FTSBrowseClansResponse",
	3105:  "EMsg_FTSSearchClansByLocation",
	3106:  "EMsg_FTSSearchClansByLocationResponse",
	3107:  "EMsg_FTSSearchPlayersByLocation",
	3108:  "EMsg_FTSSearchPlayersByLocationResponse",
	3109:  "EMsg_FTSClanDeleted",
	3110:  "EMsg_FTSSearch",
	3111:  "EMsg_FTSSearchResponse",
	3112:  "EMsg_FTSSearchStatus",
	3113:  "EMsg_FTSSearchStatusResponse",
	3114:  "EMsg_FTSGetGSPlayStats",
	3115:  "EMsg_FTSGetGSPlayStatsResponse",
	3116:  "EMsg_FTSGetGSPlayStatsForServer",
	3117:  "EMsg_FTSGetGSPlayStatsForServerResponse",
	3118:  "EMsg_FTSReportIPUpdates",
	3150:  "EMsg_BaseCCSRange",
	3151:  "EMsg_CCSGetComments",
	3152:  "EMsg_CCSGetCommentsResponse",
	3153:  "EMsg_CCSAddComment",
	3154:  "EMsg_CCSAddCommentResponse",
	3155:  "EMsg_CCSDeleteComment",
	3156:  "EMsg_CCSDeleteCommentResponse",
	3157:  "EMsg_CCSPreloadComments",
	3158:  "EMsg_CCSNotifyCommentCount",
	3159:  "EMsg_CCSGetCommentsForNews",
	3160:  "EMsg_CCSGetCommentsForNewsResponse",
	3161:  "EMsg_CCSDeleteAllCommentsByAuthor",
	3162:  "EMsg_CCSDeleteAllCommentsByAuthorResponse",
	3200:  "EMsg_BaseLBSRange",
	3201:  "EMsg_LBSSetScore",
	3202:  "EMsg_LBSSetScoreResponse",
	3203:  "EMsg_LBSFindOrCreateLB",
	3204:  "EMsg_LBSFindOrCreateLBResponse",
	3205:  "EMsg_LBSGetLBEntries",
	3206:  "EMsg_LBSGetLBEntriesResponse",
	3207:  "EMsg_LBSGetLBList",
	3208:  "EMsg_LBSGetLBListResponse",
	3209:  "EMsg_LBSSetLBDetails",
	3210:  "EMsg_LBSDeleteLB",
	3211:  "EMsg_LBSDeleteLBEntry",
	3212:  "EMsg_LBSResetLB",
	3213:  "EMsg_LBSResetLBResponse",
	3214:  "EMsg_LBSDeleteLBResponse",
	3400:  "EMsg_BaseOGS",
	3401:  "EMsg_OGSBeginSession",
	3402:  "EMsg_OGSBeginSessionResponse",
	3403:  "EMsg_OGSEndSession",
	3404:  "EMsg_OGSEndSessionResponse",
	3406:  "EMsg_OGSWriteAppSessionRow",
	3600:  "EMsg_BaseBRP",
	3601:  "EMsg_BRPStartShippingJobs",
	3602:  "EMsg_BRPProcessUSBankReports",
	3603:  "EMsg_BRPProcessGCReports",
	3604:  "EMsg_BRPProcessPPReports",
	3605:  "EMsg_BRPSettleNOVA",
	3606:  "EMsg_BRPSettleCB",
	3607:  "EMsg_BRPCommitGC",
	3608:  "EMsg_BRPCommitGCResponse",
	3609:  "EMsg_BRPFindHungTransactions",
	3610:  "EMsg_BRPCheckFinanceCloseOutDate",
	3611:  "EMsg_BRPProcessLicenses",
	3612:  "EMsg_BRPProcessLicensesResponse",
	3613:  "EMsg_BRPRemoveExpiredPaymentData",
	3614:  "EMsg_BRPRemoveExpiredPaymentDataResponse",
	3615:  "EMsg_BRPConvertToCurrentKeys",
	3616:  "EMsg_BRPConvertToCurrentKeysResponse",
	3617:  "EMsg_BRPPruneCardUsageStats",
	3618:  "EMsg_BRPPruneCardUsageStatsResponse",
	3619:  "EMsg_BRPCheckActivationCodes",
	3620:  "EMsg_BRPCheckActivationCodesResponse",
	3621:  "EMsg_BRPCommitWP",
	3622:  "EMsg_BRPCommitWPResponse",
	3623:  "EMsg_BRPProcessWPReports",
	3624:  "EMsg_BRPProcessPaymentRules",
	3625:  "EMsg_BRPProcessPartnerPayments",
	3626:  "EMsg_BRPCheckSettlementReports",
	3628:  "EMsg_BRPPostTaxToAvalara",
	3629:  "EMsg_BRPPostTransactionTax",
	3630:  "EMsg_BRPPostTransactionTaxResponse",
	3631:  "EMsg_BRPProcessIMReports",
	4000:  "EMsg_BaseAMRange2",
	4001:  "EMsg_AMCreateChat",
	4002:  "EMsg_AMCreateChatResponse",
	4003:  "EMsg_AMUpdateChatMetadata",
	4004:  "EMsg_AMPublishChatMetadata",
	4005:  "EMsg_AMSetProfileURL",
	4006:  "EMsg_AMGetAccountEmailAddress",
	4007:  "EMsg_AMGetAccountEmailAddressResponse",
	4008:  "EMsg_AMRequestFriendData",
	4009:  "EMsg_AMRouteToClients",
	4010:  "EMsg_AMLeaveClan",
	4011:  "EMsg_AMClanPermissions",
	4012:  "EMsg_AMClanPermissionsResponse",
	4013:  "EMsg_AMCreateClanEvent",
	4014:  "EMsg_AMCreateClanEventResponse",
	4015:  "EMsg_AMUpdateClanEvent",
	4016:  "EMsg_AMUpdateClanEventResponse",
	4017:  "EMsg_AMGetClanEvents",
	4018:  "EMsg_AMGetClanEventsResponse",
	4019:  "EMsg_AMDeleteClanEvent",
	4020:  "EMsg_AMDeleteClanEventResponse",
	4021:  "EMsg_AMSetClanPermissionSettings",
	4022:  "EMsg_AMSetClanPermissionSettingsResponse",
	4023:  "EMsg_AMGetClanPermissionSettings",
	4024:  "EMsg_AMGetClanPermissionSettingsResponse",
	4025:  "EMsg_AMPublishChatRoomInfo",
	4026:  "EMsg_ClientChatRoomInfo",
	4027:  "EMsg_AMCreateClanAnnouncement",
	4028:  "EMsg_AMCreateClanAnnouncementResponse",
	4029:  "EMsg_AMUpdateClanAnnouncement",
	4030:  "EMsg_AMUpdateClanAnnouncementResponse",
	4031:  "EMsg_AMGetClanAnnouncementsCount",
	4032:  "EMsg_AMGetClanAnnouncementsCountResponse",
	4033:  "EMsg_AMGetClanAnnouncements",
	4034:  "EMsg_AMGetClanAnnouncementsResponse",
	4035:  "EMsg_AMDeleteClanAnnouncement",
	4036:  "EMsg_AMDeleteClanAnnouncementResponse",
	4037:  "EMsg_AMGetSingleClanAnnouncement",
	4038:  "EMsg_AMGetSingleClanAnnouncementResponse",
	4039:  "EMsg_AMGetClanHistory",
	4040:  "EMsg_AMGetClanHistoryResponse",
	4041:  "EMsg_AMGetClanPermissionBits",
	4042:  "EMsg_AMGetClanPermissionBitsResponse",
	4043:  "EMsg_AMSetClanPermissionBits",
	4044:  "EMsg_AMSetClanPermissionBitsResponse",
	4045:  "EMsg_AMSessionInfoRequest",
	4046:  "EMsg_AMSessionInfoResponse",
	4047:  "EMsg_AMValidateWGToken",
	4048:  "EMsg_AMGetSingleClanEvent",
	4049:  "EMsg_AMGetSingleClanEventResponse",
	4050:  "EMsg_AMGetClanRank",
	4051:  "EMsg_AMGetClanRankResponse",
	4052:  "EMsg_AMSetClanRank",
	4053:  "EMsg_AMSetClanRankResponse",
	4054:  "EMsg_AMGetClanPOTW",
	4055:  "EMsg_AMGetClanPOTWResponse",
	4056:  "EMsg_AMSetClanPOTW",
	4057:  "EMsg_AMSetClanPOTWResponse",
	4058:  "EMsg_AMRequestChatMetadata",
	4059:  "EMsg_AMDumpUser",
	4060:  "EMsg_AMKickUserFromClan",
	4061:  "EMsg_AMAddFounderToClan",
	4062:  "EMsg_AMValidateWGTokenResponse",
	4063:  "EMsg_AMSetCommunityState",
	4064:  "EMsg_AMSetAccountDetails",
	4065:  "EMsg_AMGetChatBanList",
	4066:  "EMsg_AMGetChatBanListResponse",
	4067:  "EMsg_AMUnBanFromChat",
	4068:  "EMsg_AMSetClanDetails",
	4069:  "EMsg_AMGetAccountLinks",
	4070:  "EMsg_AMGetAccountLinksResponse",
	4071:  "EMsg_AMSetAccountLinks",
	4072:  "EMsg_AMSetAccountLinksResponse",
	4073:  "EMsg_AMGetUserGameStats",
	4074:  "EMsg_AMGetUserGameStatsResponse",
	4075:  "EMsg_AMCheckClanMembership",
	4076:  "EMsg_AMGetClanMembers",
	4077:  "EMsg_AMGetClanMembersResponse",
	4078:  "EMsg_AMJoinPublicClan",
	4079:  "EMsg_AMNotifyChatOfClanChange",
	4080:  "EMsg_AMResubmitPurchase",
	4081:  "EMsg_AMAddFriend",
	4082:  "EMsg_AMAddFriendResponse",
	4083:  "EMsg_AMRemoveFriend",
	4084:  "EMsg_AMDumpClan",
	4085:  "EMsg_AMChangeClanOwner",
	4086:  "EMsg_AMCancelEasyCollect",
	4087:  "EMsg_AMCancelEasyCollectResponse",
	4088:  "EMsg_AMGetClanMembershipList",
	4089:  "EMsg_AMGetClanMembershipListResponse",
	4090:  "EMsg_AMClansInCommon",
	4091:  "EMsg_AMClansInCommonResponse",
	4092:  "EMsg_AMIsValidAccountID",
	4093:  "EMsg_AMConvertClan",
	4094:  "EMsg_AMGetGiftTargetListRelay",
	4095:  "EMsg_AMWipeFriendsList",
	4096:  "EMsg_AMSetIgnored",
	4097:  "EMsg_AMClansInCommonCountResponse",
	4098:  "EMsg_AMFriendsList",
	4099:  "EMsg_AMFriendsListResponse",
	4100:  "EMsg_AMFriendsInCommon",
	4101:  "EMsg_AMFriendsInCommonResponse",
	4102:  "EMsg_AMFriendsInCommonCountResponse",
	4103:  "EMsg_AMClansInCommonCount",
	4104:  "EMsg_AMChallengeVerdict",
	4105:  "EMsg_AMChallengeNotification",
	4106:  "EMsg_AMFindGSByIP",
	4107:  "EMsg_AMFoundGSByIP",
	4108:  "EMsg_AMGiftRevoked",
	4109:  "EMsg_AMCreateAccountRecord",
	4110:  "EMsg_AMUserClanList",
	4111:  "EMsg_AMUserClanListResponse",
	4112:  "EMsg_AMGetAccountDetails2",
	4113:  "EMsg_AMGetAccountDetailsResponse2",
	4114:  "EMsg_AMSetCommunityProfileSettings",
	4115:  "EMsg_AMSetCommunityProfileSettingsResponse",
	4116:  "EMsg_AMGetCommunityPrivacyState",
	4117:  "EMsg_AMGetCommunityPrivacyStateResponse",
	4118:  "EMsg_AMCheckClanInviteRateLimiting",
	4119:  "EMsg_AMGetUserAchievementStatus",
	4120:  "EMsg_AMGetIgnored",
	4121:  "EMsg_AMGetIgnoredResponse",
	4122:  "EMsg_AMSetIgnoredResponse",
	4123:  "EMsg_AMSetFriendRelationshipNone",
	4124:  "EMsg_AMGetFriendRelationship",
	4125:  "EMsg_AMGetFriendRelationshipResponse",
	4126:  "EMsg_AMServiceModulesCache",
	4127:  "EMsg_AMServiceModulesCall",
	4128:  "EMsg_AMServiceModulesCallResponse",
	4129:  "EMsg_AMGetCaptchaDataForIP",
	4130:  "EMsg_AMGetCaptchaDataForIPResponse",
	4131:  "EMsg_AMValidateCaptchaDataForIP",
	4132:  "EMsg_AMValidateCaptchaDataForIPResponse",
	4133:  "EMsg_AMTrackFailedAuthByIP",
	4134:  "EMsg_AMGetCaptchaDataByGID",
	4135:  "EMsg_AMGetCaptchaDataByGIDResponse",
	4136:  "EMsg_AMGetLobbyList",
	4137:  "EMsg_AMGetLobbyListResponse",
	4138:  "EMsg_AMGetLobbyMetadata",
	4139:  "EMsg_AMGetLobbyMetadataResponse",
	4140:  "EMsg_CommunityAddFriendNews",
	4141:  "EMsg_AMAddClanNews",
	4142:  "EMsg_AMWriteNews",
	4143:  "EMsg_AMFindClanUser",
	4144:  "EMsg_AMFindClanUserResponse",
	4145:  "EMsg_AMBanFromChat",
	4146:  "EMsg_AMGetUserHistoryResponse",
	4147:  "EMsg_AMGetUserNewsSubscriptions",
	4148:  "EMsg_AMGetUserNewsSubscriptionsResponse",
	4149:  "EMsg_AMSetUserNewsSubscriptions",
	4150:  "EMsg_AMGetUserNews",
	4151:  "EMsg_AMGetUserNewsResponse",
	4152:  "EMsg_AMSendQueuedEmails",
	4153:  "EMsg_AMSetLicenseFlags",
	4154:  "EMsg_AMGetUserHistory",
	4155:  "EMsg_CommunityDeleteUserNews",
	4156:  "EMsg_AMAllowUserFilesRequest",
	4157:  "EMsg_AMAllowUserFilesResponse",
	4158:  "EMsg_AMGetAccountStatus",
	4159:  "EMsg_AMGetAccountStatusResponse",
	4160:  "EMsg_AMEditBanReason",
	4161:  "EMsg_AMCheckClanMembershipResponse",
	4162:  "EMsg_AMProbeClanMembershipList",
	4163:  "EMsg_AMProbeClanMembershipListResponse",
	4164:  "EMsg_UGSGetUserAchievementStatusResponse",
	4165:  "EMsg_AMGetFriendsLobbies",
	4166:  "EMsg_AMGetFriendsLobbiesResponse",
	4172:  "EMsg_AMGetUserFriendNewsResponse",
	4173:  "EMsg_CommunityGetUserFriendNews",
	4174:  "EMsg_AMGetUserClansNewsResponse",
	4175:  "EMsg_AMGetUserClansNews",
	4176:  "EMsg_AMStoreInitPurchase",
	4177:  "EMsg_AMStoreInitPurchaseResponse",
	4178:  "EMsg_AMStoreGetFinalPrice",
	4179:  "EMsg_AMStoreGetFinalPriceResponse",
	4180:  "EMsg_AMStoreCompletePurchase",
	4181:  "EMsg_AMStoreCancelPurchase",
	4182:  "EMsg_AMStorePurchaseResponse",
	4183:  "EMsg_AMCreateAccountRecordInSteam3",
	4184:  "EMsg_AMGetPreviousCBAccount",
	4185:  "EMsg_AMGetPreviousCBAccountResponse",
	4186:  "EMsg_AMUpdateBillingAddress",
	4187:  "EMsg_AMUpdateBillingAddressResponse",
	4188:  "EMsg_AMGetBillingAddress",
	4189:  "EMsg_AMGetBillingAddressResponse",
	4190:  "EMsg_AMGetUserLicenseHistory",
	4191:  "EMsg_AMGetUserLicenseHistoryResponse",
	4194:  "EMsg_AMSupportChangePassword",
	4195:  "EMsg_AMSupportChangeEmail",
	4196:  "EMsg_AMSupportChangeSecretQA",
	4197:  "EMsg_AMResetUserVerificationGSByIP",
	4198:  "EMsg_AMUpdateGSPlayStats",
	4199:  "EMsg_AMSupportEnableOrDisable",
	4200:  "EMsg_AMGetComments",
	4201:  "EMsg_AMGetCommentsResponse",
	4202:  "EMsg_AMAddComment",
	4203:  "EMsg_AMAddCommentResponse",
	4204:  "EMsg_AMDeleteComment",
	4205:  "EMsg_AMDeleteCommentResponse",
	4206:  "EMsg_AMGetPurchaseStatus",
	4209:  "EMsg_AMSupportIsAccountEnabled",
	4210:  "EMsg_AMSupportIsAccountEnabledResponse",
	4211:  "EMsg_AMGetUserStats",
	4212:  "EMsg_AMSupportKickSession",
	4213:  "EMsg_AMGSSearch",
	4216:  "EMsg_MarketingMessageUpdate",
	4219:  "EMsg_AMRouteFriendMsg",
	4220:  "EMsg_AMTicketAuthRequestOrResponse",
	4222:  "EMsg_AMVerifyDepotManagementRights",
	4223:  "EMsg_AMVerifyDepotManagementRightsResponse",
	4224:  "EMsg_AMAddFreeLicense",
	4225:  "EMsg_AMGetUserFriendsMinutesPlayed",
	4226:  "EMsg_AMGetUserFriendsMinutesPlayedResponse",
	4227:  "EMsg_AMGetUserMinutesPlayed",
	4228:  "EMsg_AMGetUserMinutesPlayedResponse",
	4231:  "EMsg_AMValidateEmailLink",
	4232:  "EMsg_AMValidateEmailLinkResponse",
	4234:  "EMsg_AMAddUsersToMarketingTreatment",
	4236:  "EMsg_AMStoreUserStats",
	4237:  "EMsg_AMGetUserGameplayInfo",
	4238:  "EMsg_AMGetUserGameplayInfoResponse",
	4239:  "EMsg_AMGetCardList",
	4240:  "EMsg_AMGetCardListResponse",
	4241:  "EMsg_AMDeleteStoredCard",
	4242:  "EMsg_AMRevokeLegacyGameKeys",
	4244:  "EMsg_AMGetWalletDetails",
	4245:  "EMsg_AMGetWalletDetailsResponse",
	4246:  "EMsg_AMDeleteStoredPaymentInfo",
	4247:  "EMsg_AMGetStoredPaymentSummary",
	4248:  "EMsg_AMGetStoredPaymentSummaryResponse",
	4249:  "EMsg_AMGetWalletConversionRate",
	4250:  "EMsg_AMGetWalletConversionRateResponse",
	4251:  "EMsg_AMConvertWallet",
	4252:  "EMsg_AMConvertWalletResponse",
	4253:  "EMsg_AMRelayGetFriendsWhoPlayGame",
	4254:  "EMsg_AMRelayGetFriendsWhoPlayGameResponse",
	4255:  "EMsg_AMSetPreApproval",
	4256:  "EMsg_AMSetPreApprovalResponse",
	4257:  "EMsg_AMMarketingTreatmentUpdate",
	4258:  "EMsg_AMCreateRefund",
	4259:  "EMsg_AMCreateRefundResponse",
	4260:  "EMsg_AMCreateChargeback",
	4261:  "EMsg_AMCreateChargebackResponse",
	4262:  "EMsg_AMCreateDispute",
	4263:  "EMsg_AMCreateDisputeResponse",
	4264:  "EMsg_AMClearDispute",
	4265:  "EMsg_AMClearDisputeResponse",
	4266:  "EMsg_AMPlayerNicknameList",
	4267:  "EMsg_AMPlayerNicknameListResponse",
	4268:  "EMsg_AMSetDRMTestConfig",
	4269:  "EMsg_AMGetUserCurrentGameInfo",
	4270:  "EMsg_AMGetUserCurrentGameInfoResponse",
	4271:  "EMsg_AMGetGSPlayerList",
	4272:  "EMsg_AMGetGSPlayerListResponse",
	4275:  "EMsg_AMUpdatePersonaStateCache",
	4276:  "EMsg_AMGetGameMembers",
	4277:  "EMsg_AMGetGameMembersResponse",
	4278:  "EMsg_AMGetSteamIDForMicroTxn",
	4279:  "EMsg_AMGetSteamIDForMicroTxnResponse",
	4280:  "EMsg_AMAddPublisherUser",
	4281:  "EMsg_AMRemovePublisherUser",
	4282:  "EMsg_AMGetUserLicenseList",
	4283:  "EMsg_AMGetUserLicenseListResponse",
	4284:  "EMsg_AMReloadGameGroupPolicy",
	4285:  "EMsg_AMAddFreeLicenseResponse",
	4286:  "EMsg_AMVACStatusUpdate",
	4287:  "EMsg_AMGetAccountDetails",
	4288:  "EMsg_AMGetAccountDetailsResponse",
	4289:  "EMsg_AMGetPlayerLinkDetails",
	4290:  "EMsg_AMGetPlayerLinkDetailsResponse",
	4291:  "EMsg_AMSubscribeToPersonaFeed",
	4292:  "EMsg_AMGetUserVacBanList",
	4293:  "EMsg_AMGetUserVacBanListResponse",
	4294:  "EMsg_AMGetAccountFlagsForWGSpoofing",
	4295:  "EMsg_AMGetAccountFlagsForWGSpoofingResponse",
	4296:  "EMsg_AMGetFriendsWishlistInfo",
	4297:  "EMsg_AMGetFriendsWishlistInfoResponse",
	4298:  "EMsg_AMGetClanOfficers",
	4299:  "EMsg_AMGetClanOfficersResponse",
	4300:  "EMsg_AMNameChange",
	4301:  "EMsg_AMGetNameHistory",
	4302:  "EMsg_AMGetNameHistoryResponse",
	4305:  "EMsg_AMUpdateProviderStatus",
	4306:  "EMsg_AMClearPersonaMetadataBlob",
	4307:  "EMsg_AMSupportRemoveAccountSecurity",
	4308:  "EMsg_AMIsAccountInCaptchaGracePeriod",
	4309:  "EMsg_AMIsAccountInCaptchaGracePeriodResponse",
	4310:  "EMsg_AMAccountPS3Unlink",
	4311:  "EMsg_AMAccountPS3UnlinkResponse",
	4312:  "EMsg_AMStoreUserStatsResponse",
	4313:  "EMsg_AMGetAccountPSNInfo",
	4314:  "EMsg_AMGetAccountPSNInfoResponse",
	4315:  "EMsg_AMAuthenticatedPlayerList",
	4316:  "EMsg_AMGetUserGifts",
	4317:  "EMsg_AMGetUserGiftsResponse",
	4320:  "EMsg_AMTransferLockedGifts",
	4321:  "EMsg_AMTransferLockedGiftsResponse",
	4322:  "EMsg_AMPlayerHostedOnGameServer",
	4323:  "EMsg_AMGetAccountBanInfo",
	4324:  "EMsg_AMGetAccountBanInfoResponse",
	4325:  "EMsg_AMRecordBanEnforcement",
	4326:  "EMsg_AMRollbackGiftTransfer",
	4327:  "EMsg_AMRollbackGiftTransferResponse",
	4328:  "EMsg_AMHandlePendingTransaction",
	4329:  "EMsg_AMRequestClanDetails",
	4330:  "EMsg_AMDeleteStoredPaypalAgreement",
	4331:  "EMsg_AMGameServerUpdate",
	4332:  "EMsg_AMGameServerRemove",
	4333:  "EMsg_AMGetPaypalAgreements",
	4334:  "EMsg_AMGetPaypalAgreementsResponse",
	4335:  "EMsg_AMGameServerPlayerCompatibilityCheck",
	4336:  "EMsg_AMGameServerPlayerCompatibilityCheckResponse",
	4337:  "EMsg_AMRenewLicense",
	4338:  "EMsg_AMGetAccountCommunityBanInfo",
	4339:  "EMsg_AMGetAccountCommunityBanInfoResponse",
	4340:  "EMsg_AMGameServerAccountChangePassword",
	4341:  "EMsg_AMGameServerAccountDeleteAccount",
	4342:  "EMsg_AMRenewAgreement",
	4343:  "EMsg_AMSendEmail",
	4344:  "EMsg_AMXsollaPayment",
	4345:  "EMsg_AMXsollaPaymentResponse",
	4346:  "EMsg_AMAcctAllowedToPurchase",
	4347:  "EMsg_AMAcctAllowedToPurchaseResponse",
	4348:  "EMsg_AMSwapKioskDeposit",
	4349:  "EMsg_AMSwapKioskDepositResponse",
	4350:  "EMsg_AMSetUserGiftUnowned",
	4351:  "EMsg_AMSetUserGiftUnownedResponse",
	4352:  "EMsg_AMClaimUnownedUserGift",
	4353:  "EMsg_AMClaimUnownedUserGiftResponse",
	4354:  "EMsg_AMSetClanName",
	4355:  "EMsg_AMSetClanNameResponse",
	4356:  "EMsg_AMGrantCoupon",
	4357:  "EMsg_AMGrantCouponResponse",
	4358:  "EMsg_AMIsPackageRestrictedInUserCountry",
	4359:  "EMsg_AMIsPackageRestrictedInUserCountryResponse",
	4360:  "EMsg_AMHandlePendingTransactionResponse",
	4361:  "EMsg_AMGrantGuestPasses2",
	4362:  "EMsg_AMGrantGuestPasses2Response",
	4363:  "EMsg_AMSessionQuery",
	4364:  "EMsg_AMSessionQueryResponse",
	4365:  "EMsg_AMGetPlayerBanDetails",
	4366:  "EMsg_AMGetPlayerBanDetailsResponse",
	4367:  "EMsg_AMFinalizePurchase",
	4368:  "EMsg_AMFinalizePurchaseResponse",
	4372:  "EMsg_AMPersonaChangeResponse",
	4373:  "EMsg_AMGetClanDetailsForForumCreation",
	4374:  "EMsg_AMGetClanDetailsForForumCreationResponse",
	4375:  "EMsg_AMGetPendingNotificationCount",
	4376:  "EMsg_AMGetPendingNotificationCountResponse",
	4377:  "EMsg_AMPasswordHashUpgrade",
	4378:  "EMsg_AMMoPayPayment",
	4379:  "EMsg_AMMoPayPaymentResponse",
	4380:  "EMsg_AMBoaCompraPayment",
	4381:  "EMsg_AMBoaCompraPaymentResponse",
	4382:  "EMsg_AMExpireCaptchaByGID",
	4383:  "EMsg_AMCompleteExternalPurchase",
	4384:  "EMsg_AMCompleteExternalPurchaseResponse",
	4385:  "EMsg_AMResolveNegativeWalletCredits",
	4386:  "EMsg_AMResolveNegativeWalletCreditsResponse",
	4387:  "EMsg_AMPayelpPayment",
	4388:  "EMsg_AMPayelpPaymentResponse",
	4389:  "EMsg_AMPlayerGetClanBasicDetails",
	4390:  "EMsg_AMPlayerGetClanBasicDetailsResponse",
	4391:  "EMsg_AMMOLPayment",
	4392:  "EMsg_AMMOLPaymentResponse",
	4393:  "EMsg_GetUserIPCountry",
	4394:  "EMsg_GetUserIPCountryResponse",
	4395:  "EMsg_NotificationOfSuspiciousActivity",
	4396:  "EMsg_AMDegicaPayment",
	4397:  "EMsg_AMDegicaPaymentResponse",
	4398:  "EMsg_AMEClubPayment",
	4399:  "EMsg_AMEClubPaymentResponse",
	4400:  "EMsg_AMPayPalPaymentsHubPayment",
	4401:  "EMsg_AMPayPalPaymentsHubPaymentResponse",
	4402:  "EMsg_AMTwoFactorRecoverAuthenticatorRequest",
	4403:  "EMsg_AMTwoFactorRecoverAuthenticatorResponse",
	4404:  "EMsg_AMSmart2PayPayment",
	4405:  "EMsg_AMSmart2PayPaymentResponse",
	4406:  "EMsg_AMValidatePasswordResetCodeAndSendSmsRequest",
	4407:  "EMsg_AMValidatePasswordResetCodeAndSendSmsResponse",
	4408:  "EMsg_AMGetAccountResetDetailsRequest",
	4409:  "EMsg_AMGetAccountResetDetailsResponse",
	4410:  "EMsg_AMBitPayPayment",
	4411:  "EMsg_AMBitPayPaymentResponse",
	4412:  "EMsg_AMSendAccountInfoUpdate",
	4413:  "EMsg_AMSendScheduledGift",
	4414:  "EMsg_AMNodwinPayment",
	4415:  "EMsg_AMNodwinPaymentResponse",
	4416:  "EMsg_AMResolveWalletRevoke",
	4417:  "EMsg_AMResolveWalletReverseRevoke",
	4418:  "EMsg_AMFundedPayment",
	4419:  "EMsg_AMFundedPaymentResponse",
	4420:  "EMsg_AMRequestPersonaUpdateForChatServer",
	4421:  "EMsg_AMPerfectWorldPayment",
	4422:  "EMsg_AMPerfectWorldPaymentResponse",
	5000:  "EMsg_BasePSRange",
	5001:  "EMsg_PSCreateShoppingCart",
	5002:  "EMsg_PSCreateShoppingCartResponse",
	5003:  "EMsg_PSIsValidShoppingCart",
	5004:  "EMsg_PSIsValidShoppingCartResponse",
	5005:  "EMsg_PSAddPackageToShoppingCart",
	5006:  "EMsg_PSAddPackageToShoppingCartResponse",
	5007:  "EMsg_PSRemoveLineItemFromShoppingCart",
	5008:  "EMsg_PSRemoveLineItemFromShoppingCartResponse",
	5009:  "EMsg_PSGetShoppingCartContents",
	5010:  "EMsg_PSGetShoppingCartContentsResponse",
	5011:  "EMsg_PSAddWalletCreditToShoppingCart",
	5012:  "EMsg_PSAddWalletCreditToShoppingCartResponse",
	5200:  "EMsg_BaseUFSRange",
	5202:  "EMsg_ClientUFSUploadFileRequest",
	5203:  "EMsg_ClientUFSUploadFileResponse",
	5204:  "EMsg_ClientUFSUploadFileChunk",
	5205:  "EMsg_ClientUFSUploadFileFinished",
	5206:  "EMsg_ClientUFSGetFileListForApp",
	5207:  "EMsg_ClientUFSGetFileListForAppResponse",
	5210:  "EMsg_ClientUFSDownloadRequest",
	5211:  "EMsg_ClientUFSDownloadResponse",
	5212:  "EMsg_ClientUFSDownloadChunk",
	5213:  "EMsg_ClientUFSLoginRequest",
	5214:  "EMsg_ClientUFSLoginResponse",
	5215:  "EMsg_UFSReloadPartitionInfo",
	5216:  "EMsg_ClientUFSTransferHeartbeat",
	5217:  "EMsg_UFSSynchronizeFile",
	5218:  "EMsg_UFSSynchronizeFileResponse",
	5219:  "EMsg_ClientUFSDeleteFileRequest",
	5220:  "EMsg_ClientUFSDeleteFileResponse",
	5221:  "EMsg_UFSDownloadRequest",
	5222:  "EMsg_UFSDownloadResponse",
	5223:  "EMsg_UFSDownloadChunk",
	5226:  "EMsg_ClientUFSGetUGCDetails",
	5227:  "EMsg_ClientUFSGetUGCDetailsResponse",
	5228:  "EMsg_UFSUpdateFileFlags",
	5229:  "EMsg_UFSUpdateFileFlagsResponse",
	5230:  "EMsg_ClientUFSGetSingleFileInfo",
	5231:  "EMsg_ClientUFSGetSingleFileInfoResponse",
	5232:  "EMsg_ClientUFSShareFile",
	5233:  "EMsg_ClientUFSShareFileResponse",
	5234:  "EMsg_UFSReloadAccount",
	5235:  "EMsg_UFSReloadAccountResponse",
	5236:  "EMsg_UFSUpdateRecordBatched",
	5237:  "EMsg_UFSUpdateRecordBatchedResponse",
	5238:  "EMsg_UFSMigrateFile",
	5239:  "EMsg_UFSMigrateFileResponse",
	5240:  "EMsg_UFSGetUGCURLs",
	5241:  "EMsg_UFSGetUGCURLsResponse",
	5242:  "EMsg_UFSHttpUploadFileFinishRequest",
	5243:  "EMsg_UFSHttpUploadFileFinishResponse",
	5244:  "EMsg_UFSDownloadStartRequest",
	5245:  "EMsg_UFSDownloadStartResponse",
	5246:  "EMsg_UFSDownloadChunkRequest",
	5247:  "EMsg_UFSDownloadChunkResponse",
	5248:  "EMsg_UFSDownloadFinishRequest",
	5249:  "EMsg_UFSDownloadFinishResponse",
	5250:  "EMsg_UFSFlushURLCache",
	5251:  "EMsg_UFSUploadCommit",
	5252:  "EMsg_UFSUploadCommitResponse",
	5253:  "EMsg_UFSMigrateFileAppID",
	5254:  "EMsg_UFSMigrateFileAppIDResponse",
	5400:  "EMsg_BaseClient2",
	5401:  "EMsg_ClientRequestForgottenPasswordEmail",
	5402:  "EMsg_ClientRequestForgottenPasswordEmailResponse",
	5403:  "EMsg_ClientCreateAccountResponse",
	5404:  "EMsg_ClientResetForgottenPassword",
	5405:  "EMsg_ClientResetForgottenPasswordResponse",
	5406:  "EMsg_ClientCreateAccount2",
	5407:  "EMsg_ClientInformOfResetForgottenPassword",
	5408:  "EMsg_ClientInformOfResetForgottenPasswordResponse",
	5409:  "EMsg_ClientAnonUserLogOn_Deprecated",
	5410:  "EMsg_ClientGamesPlayedWithDataBlob",
	5411:  "EMsg_ClientUpdateUserGameInfo",
	5412:  "EMsg_ClientFileToDownload",
	5413:  "EMsg_ClientFileToDownloadResponse",
	5414:  "EMsg_ClientLBSSetScore",
	5415:  "EMsg_ClientLBSSetScoreResponse",
	5416:  "EMsg_ClientLBSFindOrCreateLB",
	5417:  "EMsg_ClientLBSFindOrCreateLBResponse",
	5418:  "EMsg_ClientLBSGetLBEntries",
	5419:  "EMsg_ClientLBSGetLBEntriesResponse",
	5420:  "EMsg_ClientMarketingMessageUpdate",
	5426:  "EMsg_ClientChatDeclined",
	5427:  "EMsg_ClientFriendMsgIncoming",
	5428:  "EMsg_ClientAuthList_Deprecated",
	5429:  "EMsg_ClientTicketAuthComplete",
	5430:  "EMsg_ClientIsLimitedAccount",
	5431:  "EMsg_ClientRequestAuthList",
	5432:  "EMsg_ClientAuthList",
	5433:  "EMsg_ClientStat",
	5434:  "EMsg_ClientP2PConnectionInfo",
	5435:  "EMsg_ClientP2PConnectionFailInfo",
	5436:  "EMsg_ClientGetNumberOfCurrentPlayers",
	5437:  "EMsg_ClientGetNumberOfCurrentPlayersResponse",
	5438:  "EMsg_ClientGetDepotDecryptionKey",
	5439:  "EMsg_ClientGetDepotDecryptionKeyResponse",
	5440:  "EMsg_GSPerformHardwareSurvey",
	5441:  "EMsg_ClientGetAppBetaPasswords",
	5442:  "EMsg_ClientGetAppBetaPasswordsResponse",
	5443:  "EMsg_ClientEnableTestLicense",
	5444:  "EMsg_ClientEnableTestLicenseResponse",
	5445:  "EMsg_ClientDisableTestLicense",
	5446:  "EMsg_ClientDisableTestLicenseResponse",
	5448:  "EMsg_ClientRequestValidationMail",
	5449:  "EMsg_ClientRequestValidationMailResponse",
	5450:  "EMsg_ClientCheckAppBetaPassword",
	5451:  "EMsg_ClientCheckAppBetaPasswordResponse",
	5452:  "EMsg_ClientToGC",
	5453:  "EMsg_ClientFromGC",
	5454:  "EMsg_ClientRequestChangeMail",
	5455:  "EMsg_ClientRequestChangeMailResponse",
	5456:  "EMsg_ClientEmailAddrInfo",
	5457:  "EMsg_ClientPasswordChange3",
	5458:  "EMsg_ClientEmailChange3",
	5459:  "EMsg_ClientPersonalQAChange3",
	5460:  "EMsg_ClientResetForgottenPassword3",
	5461:  "EMsg_ClientRequestForgottenPasswordEmail3",
	5462:  "EMsg_ClientCreateAccount3",
	5463:  "EMsg_ClientNewLoginKey",
	5464:  "EMsg_ClientNewLoginKeyAccepted",
	5465:  "EMsg_ClientLogOnWithHash_Deprecated",
	5466:  "EMsg_ClientStoreUserStats2",
	5467:  "EMsg_ClientStatsUpdated",
	5468:  "EMsg_ClientActivateOEMLicense",
	5469:  "EMsg_ClientRegisterOEMMachine",
	5470:  "EMsg_ClientRegisterOEMMachineResponse",
	5480:  "EMsg_ClientRequestedClientStats",
	5481:  "EMsg_ClientStat2Int32",
	5482:  "EMsg_ClientStat2",
	5483:  "EMsg_ClientVerifyPassword",
	5484:  "EMsg_ClientVerifyPasswordResponse",
	5485:  "EMsg_ClientDRMDownloadRequest",
	5486:  "EMsg_ClientDRMDownloadResponse",
	5487:  "EMsg_ClientDRMFinalResult",
	5488:  "EMsg_ClientGetFriendsWhoPlayGame",
	5489:  "EMsg_ClientGetFriendsWhoPlayGameResponse",
	5490:  "EMsg_ClientOGSBeginSession",
	5491:  "EMsg_ClientOGSBeginSessionResponse",
	5492:  "EMsg_ClientOGSEndSession",
	5493:  "EMsg_ClientOGSEndSessionResponse",
	5494:  "EMsg_ClientOGSWriteRow",
	5495:  "EMsg_ClientDRMTest",
	5496:  "EMsg_ClientDRMTestResult",
	5500:  "EMsg_ClientServerUnavailable",
	5501:  "EMsg_ClientServersAvailable",
	5502:  "EMsg_ClientRegisterAuthTicketWithCM",
	5503:  "EMsg_ClientGCMsgFailed",
	5504:  "EMsg_ClientMicroTxnAuthRequest",
	5505:  "EMsg_ClientMicroTxnAuthorize",
	5506:  "EMsg_ClientMicroTxnAuthorizeResponse",
	5507:  "EMsg_ClientAppMinutesPlayedData",
	5508:  "EMsg_ClientGetMicroTxnInfo",
	5509:  "EMsg_ClientGetMicroTxnInfoResponse",
	5510:  "EMsg_ClientMarketingMessageUpdate2",
	5511:  "EMsg_ClientDeregisterWithServer",
	5512:  "EMsg_ClientSubscribeToPersonaFeed",
	5514:  "EMsg_ClientLogon",
	5515:  "EMsg_ClientGetClientDetails",
	5516:  "EMsg_ClientGetClientDetailsResponse",
	5517:  "EMsg_ClientReportOverlayDetourFailure",
	5518:  "EMsg_ClientGetClientAppList",
	5519:  "EMsg_ClientGetClientAppListResponse",
	5520:  "EMsg_ClientInstallClientApp",
	5521:  "EMsg_ClientInstallClientAppResponse",
	5522:  "EMsg_ClientUninstallClientApp",
	5523:  "EMsg_ClientUninstallClientAppResponse",
	5524:  "EMsg_ClientSetClientAppUpdateState",
	5525:  "EMsg_ClientSetClientAppUpdateStateResponse",
	5526:  "EMsg_ClientRequestEncryptedAppTicket",
	5527:  "EMsg_ClientRequestEncryptedAppTicketResponse",
	5528:  "EMsg_ClientWalletInfoUpdate",
	5529:  "EMsg_ClientLBSSetUGC",
	5530:  "EMsg_ClientLBSSetUGCResponse",
	5531:  "EMsg_ClientAMGetClanOfficers",
	5532:  "EMsg_ClientAMGetClanOfficersResponse",
	5533:  "EMsg_ClientCheckFileSignature",
	5534:  "EMsg_ClientCheckFileSignatureResponse",
	5535:  "EMsg_ClientFriendProfileInfo",
	5536:  "EMsg_ClientFriendProfileInfoResponse",
	5537:  "EMsg_ClientUpdateMachineAuth",
	5538:  "EMsg_ClientUpdateMachineAuthResponse",
	5539:  "EMsg_ClientReadMachineAuth",
	5540:  "EMsg_ClientReadMachineAuthResponse",
	5541:  "EMsg_ClientRequestMachineAuth",
	5542:  "EMsg_ClientRequestMachineAuthResponse",
	5543:  "EMsg_ClientScreenshotsChanged",
	5544:  "EMsg_ClientEmailChange4",
	5545:  "EMsg_ClientEmailChangeResponse4",
	5546:  "EMsg_ClientGetCDNAuthToken",
	5547:  "EMsg_ClientGetCDNAuthTokenResponse",
	5548:  "EMsg_ClientDownloadRateStatistics",
	5549:  "EMsg_ClientRequestAccountData",
	5550:  "EMsg_ClientRequestAccountDataResponse",
	5551:  "EMsg_ClientResetForgottenPassword4",
	5552:  "EMsg_ClientHideFriend",
	5553:  "EMsg_ClientFriendsGroupsList",
	5554:  "EMsg_ClientGetClanActivityCounts",
	5555:  "EMsg_ClientGetClanActivityCountsResponse",
	5556:  "EMsg_ClientOGSReportString",
	5557:  "EMsg_ClientOGSReportBug",
	5558:  "EMsg_ClientSentLogs",
	5559:  "EMsg_ClientLogonGameServer",
	5560:  "EMsg_AMClientCreateFriendsGroup",
	5561:  "EMsg_AMClientCreateFriendsGroupResponse",
	5562:  "EMsg_AMClientDeleteFriendsGroup",
	5563:  "EMsg_AMClientDeleteFriendsGroupResponse",
	5564:  "EMsg_AMClientRenameFriendsGroup",
	5565:  "EMsg_AMClientRenameFriendsGroupResponse",
	5566:  "EMsg_AMClientAddFriendToGroup",
	5567:  "EMsg_AMClientAddFriendToGroupResponse",
	5568:  "EMsg_AMClientRemoveFriendFromGroup",
	5569:  "EMsg_AMClientRemoveFriendFromGroupResponse",
	5570:  "EMsg_ClientAMGetPersonaNameHistory",
	5571:  "EMsg_ClientAMGetPersonaNameHistoryResponse",
	5572:  "EMsg_ClientRequestFreeLicense",
	5573:  "EMsg_ClientRequestFreeLicenseResponse",
	5574:  "EMsg_ClientDRMDownloadRequestWithCrashData",
	5575:  "EMsg_ClientAuthListAck",
	5576:  "EMsg_ClientItemAnnouncements",
	5577:  "EMsg_ClientRequestItemAnnouncements",
	5578:  "EMsg_ClientFriendMsgEchoToSender",
	5579:  "EMsg_ClientChangeSteamGuardOptions",
	5580:  "EMsg_ClientChangeSteamGuardOptionsResponse",
	5581:  "EMsg_ClientOGSGameServerPingSample",
	5582:  "EMsg_ClientCommentNotifications",
	5583:  "EMsg_ClientRequestCommentNotifications",
	5584:  "EMsg_ClientPersonaChangeResponse",
	5585:  "EMsg_ClientRequestWebAPIAuthenticateUserNonce",
	5586:  "EMsg_ClientRequestWebAPIAuthenticateUserNonceResponse",
	5587:  "EMsg_ClientPlayerNicknameList",
	5588:  "EMsg_AMClientSetPlayerNickname",
	5589:  "EMsg_AMClientSetPlayerNicknameResponse",
	5590:  "EMsg_ClientRequestOAuthTokenForApp",
	5591:  "EMsg_ClientRequestOAuthTokenForAppResponse",
	5592:  "EMsg_ClientGetNumberOfCurrentPlayersDP",
	5593:  "EMsg_ClientGetNumberOfCurrentPlayersDPResponse",
	5594:  "EMsg_ClientServiceMethod",
	5595:  "EMsg_ClientServiceMethodResponse",
	5596:  "EMsg_ClientFriendUserStatusPublished",
	5597:  "EMsg_ClientCurrentUIMode",
	5598:  "EMsg_ClientVanityURLChangedNotification",
	5599:  "EMsg_ClientUserNotifications",
	5600:  "EMsg_BaseDFS",
	5601:  "EMsg_DFSGetFile",
	5602:  "EMsg_DFSInstallLocalFile",
	5603:  "EMsg_DFSConnection",
	5604:  "EMsg_DFSConnectionReply",
	5605:  "EMsg_ClientDFSAuthenticateRequest",
	5606:  "EMsg_ClientDFSAuthenticateResponse",
	5607:  "EMsg_ClientDFSEndSession",
	5608:  "EMsg_DFSPurgeFile",
	5609:  "EMsg_DFSRouteFile",
	5610:  "EMsg_DFSGetFileFromServer",
	5611:  "EMsg_DFSAcceptedResponse",
	5612:  "EMsg_DFSRequestPingback",
	5613:  "EMsg_DFSRecvTransmitFile",
	5614:  "EMsg_DFSSendTransmitFile",
	5615:  "EMsg_DFSRequestPingback2",
	5616:  "EMsg_DFSResponsePingback2",
	5617:  "EMsg_ClientDFSDownloadStatus",
	5618:  "EMsg_DFSStartTransfer",
	5619:  "EMsg_DFSTransferComplete",
	5620:  "EMsg_DFSRouteFileResponse",
	5621:  "EMsg_ClientNetworkingCertRequest",
	5622:  "EMsg_ClientNetworkingCertRequestResponse",
	5623:  "EMsg_ClientChallengeRequest",
	5624:  "EMsg_ClientChallengeResponse",
	5625:  "EMsg_BadgeCraftedNotification",
	5626:  "EMsg_ClientNetworkingMobileCertRequest",
	5627:  "EMsg_ClientNetworkingMobileCertRequestResponse",
	5800:  "EMsg_BaseMDS",
	5801:  "EMsg_ClientMDSLoginRequest",
	5802:  "EMsg_ClientMDSLoginResponse",
	5803:  "EMsg_ClientMDSUploadManifestRequest",
	5804:  "EMsg_ClientMDSUploadManifestResponse",
	5805:  "EMsg_ClientMDSTransmitManifestDataChunk",
	5806:  "EMsg_ClientMDSHeartbeat",
	5807:  "EMsg_ClientMDSUploadDepotChunks",
	5808:  "EMsg_ClientMDSUploadDepotChunksResponse",
	5809:  "EMsg_ClientMDSInitDepotBuildRequest",
	5810:  "EMsg_ClientMDSInitDepotBuildResponse",
	5812:  "EMsg_AMToMDSGetDepotDecryptionKey",
	5813:  "EMsg_MDSToAMGetDepotDecryptionKeyResponse",
	5814:  "EMsg_MDSGetVersionsForDepot",
	5815:  "EMsg_MDSGetVersionsForDepotResponse",
	5816:  "EMsg_MDSSetPublicVersionForDepot",
	5817:  "EMsg_MDSSetPublicVersionForDepotResponse",
	5818:  "EMsg_ClientMDSGetDepotManifest",
	5819:  "EMsg_ClientMDSGetDepotManifestResponse",
	5820:  "EMsg_ClientMDSGetDepotManifestChunk",
	5823:  "EMsg_ClientMDSUploadRateTest",
	5824:  "EMsg_ClientMDSUploadRateTestResponse",
	5825:  "EMsg_MDSDownloadDepotChunksAck",
	5826:  "EMsg_MDSContentServerStatsBroadcast",
	5827:  "EMsg_MDSContentServerConfigRequest",
	5828:  "EMsg_MDSContentServerConfig",
	5829:  "EMsg_MDSGetDepotManifest",
	5830:  "EMsg_MDSGetDepotManifestResponse",
	5831:  "EMsg_MDSGetDepotManifestChunk",
	5832:  "EMsg_MDSGetDepotChunk",
	5833:  "EMsg_MDSGetDepotChunkResponse",
	5834:  "EMsg_MDSGetDepotChunkChunk",
	5835:  "EMsg_MDSUpdateContentServerConfig",
	5836:  "EMsg_MDSGetServerListForUser",
	5837:  "EMsg_MDSGetServerListForUserResponse",
	5838:  "EMsg_ClientMDSRegisterAppBuild",
	5839:  "EMsg_ClientMDSRegisterAppBuildResponse",
	5840:  "EMsg_ClientMDSSetAppBuildLive",
	5841:  "EMsg_ClientMDSSetAppBuildLiveResponse",
	5842:  "EMsg_ClientMDSGetPrevDepotBuild",
	5843:  "EMsg_ClientMDSGetPrevDepotBuildResponse",
	5844:  "EMsg_MDSToCSFlushChunk",
	5845:  "EMsg_ClientMDSSignInstallScript",
	5846:  "EMsg_ClientMDSSignInstallScriptResponse",
	5847:  "EMsg_MDSMigrateChunk",
	5848:  "EMsg_MDSMigrateChunkResponse",
	5849:  "EMsg_MDSToCSFlushManifest",
	6200:  "EMsg_CSBase",
	6201:  "EMsg_CSPing",
	6202:  "EMsg_CSPingResponse",
	6400:  "EMsg_GMSBase",
	6401:  "EMsg_GMSGameServerReplicate",
	6403:  "EMsg_ClientGMSServerQuery",
	6404:  "EMsg_GMSClientServerQueryResponse",
	6405:  "EMsg_AMGMSGameServerUpdate",
	6406:  "EMsg_AMGMSGameServerRemove",
	6407:  "EMsg_GameServerOutOfDate",
	6500:  "EMsg_DeviceAuthorizationBase",
	6501:  "EMsg_ClientAuthorizeLocalDeviceRequest",
	6502:  "EMsg_ClientAuthorizeLocalDevice",
	6503:  "EMsg_ClientDeauthorizeDeviceRequest",
	6504:  "EMsg_ClientDeauthorizeDevice",
	6505:  "EMsg_ClientUseLocalDeviceAuthorizations",
	6506:  "EMsg_ClientGetAuthorizedDevices",
	6507:  "EMsg_ClientGetAuthorizedDevicesResponse",
	6508:  "EMsg_AMNotifySessionDeviceAuthorized",
	6509:  "EMsg_ClientAuthorizeLocalDeviceNotification",
	6600:  "EMsg_MMSBase",
	6601:  "EMsg_ClientMMSCreateLobby",
	6602:  "EMsg_ClientMMSCreateLobbyResponse",
	6603:  "EMsg_ClientMMSJoinLobby",
	6604:  "EMsg_ClientMMSJoinLobbyResponse",
	6605:  "EMsg_ClientMMSLeaveLobby",
	6606:  "EMsg_ClientMMSLeaveLobbyResponse",
	6607:  "EMsg_ClientMMSGetLobbyList",
	6608:  "EMsg_ClientMMSGetLobbyListResponse",
	6609:  "EMsg_ClientMMSSetLobbyData",
	6610:  "EMsg_ClientMMSSetLobbyDataResponse",
	6611:  "EMsg_ClientMMSGetLobbyData",
	6612:  "EMsg_ClientMMSLobbyData",
	6613:  "EMsg_ClientMMSSendLobbyChatMsg",
	6614:  "EMsg_ClientMMSLobbyChatMsg",
	6615:  "EMsg_ClientMMSSetLobbyOwner",
	6616:  "EMsg_ClientMMSSetLobbyOwnerResponse",
	6617:  "EMsg_ClientMMSSetLobbyGameServer",
	6618:  "EMsg_ClientMMSLobbyGameServerSet",
	6619:  "EMsg_ClientMMSUserJoinedLobby",
	6620:  "EMsg_ClientMMSUserLeftLobby",
	6621:  "EMsg_ClientMMSInviteToLobby",
	6622:  "EMsg_ClientMMSFlushFrenemyListCache",
	6623:  "EMsg_ClientMMSFlushFrenemyListCacheResponse",
	6624:  "EMsg_ClientMMSSetLobbyLinked",
	6625:  "EMsg_ClientMMSSetRatelimitPolicyOnClient",
	6626:  "EMsg_ClientMMSGetLobbyStatus",
	6627:  "EMsg_ClientMMSGetLobbyStatusResponse",
	6628:  "EMsg_MMSGetLobbyList",
	6629:  "EMsg_MMSGetLobbyListResponse",
	6800:  "EMsg_NonStdMsgBase",
	6801:  "EMsg_NonStdMsgMemcached",
	6802:  "EMsg_NonStdMsgHTTPServer",
	6803:  "EMsg_NonStdMsgHTTPClient",
	6804:  "EMsg_NonStdMsgWGResponse",
	6805:  "EMsg_NonStdMsgPHPSimulator",
	6806:  "EMsg_NonStdMsgChase",
	6807:  "EMsg_NonStdMsgDFSTransfer",
	6808:  "EMsg_NonStdMsgTests",
	6809:  "EMsg_NonStdMsgUMQpipeAAPL",
	6810:  "EMsg_NonStdMsgSyslog",
	6811:  "EMsg_NonStdMsgLogsink",
	6812:  "EMsg_NonStdMsgSteam2Emulator",
	6813:  "EMsg_NonStdMsgRTMPServer",
	6814:  "EMsg_NonStdMsgWebSocket",
	6815:  "EMsg_NonStdMsgRedis",
	7000:  "EMsg_UDSBase",
	7001:  "EMsg_ClientUDSP2PSessionStarted",
	7002:  "EMsg_ClientUDSP2PSessionEnded",
	7003:  "EMsg_UDSRenderUserAuth",
	7004:  "EMsg_UDSRenderUserAuthResponse",
	7005:  "EMsg_ClientUDSInviteToGame",
	7006:  "EMsg_UDSFindSession",
	7007:  "EMsg_UDSFindSessionResponse",
	7100:  "EMsg_MPASBase",
	7101:  "EMsg_MPASVacBanReset",
	7200:  "EMsg_KGSBase",
	7201:  "EMsg_KGSAllocateKeyRange",
	7202:  "EMsg_KGSAllocateKeyRangeResponse",
	7203:  "EMsg_KGSGenerateKeys",
	7204:  "EMsg_KGSGenerateKeysResponse",
	7205:  "EMsg_KGSRemapKeys",
	7206:  "EMsg_KGSRemapKeysResponse",
	7207:  "EMsg_KGSGenerateGameStopWCKeys",
	7208:  "EMsg_KGSGenerateGameStopWCKeysResponse",
	7300:  "EMsg_UCMBase",
	7301:  "EMsg_ClientUCMAddScreenshot",
	7302:  "EMsg_ClientUCMAddScreenshotResponse",
	7303:  "EMsg_UCMValidateObjectExists",
	7304:  "EMsg_UCMValidateObjectExistsResponse",
	7307:  "EMsg_UCMResetCommunityContent",
	7308:  "EMsg_UCMResetCommunityContentResponse",
	7309:  "EMsg_ClientUCMDeleteScreenshot",
	7310:  "EMsg_ClientUCMDeleteScreenshotResponse",
	7311:  "EMsg_ClientUCMPublishFile",
	7312:  "EMsg_ClientUCMPublishFileResponse",
	7313:  "EMsg_ClientUCMGetPublishedFileDetails",
	7314:  "EMsg_ClientUCMGetPublishedFileDetailsResponse",
	7315:  "EMsg_ClientUCMDeletePublishedFile",
	7316:  "EMsg_ClientUCMDeletePublishedFileResponse",
	7317:  "EMsg_ClientUCMEnumerateUserPublishedFiles",
	7318:  "EMsg_ClientUCMEnumerateUserPublishedFilesResponse",
	7319:  "EMsg_ClientUCMSubscribePublishedFile",
	7320:  "EMsg_ClientUCMSubscribePublishedFileResponse",
	7321:  "EMsg_ClientUCMEnumerateUserSubscribedFiles",
	7322:  "EMsg_ClientUCMEnumerateUserSubscribedFilesResponse",
	7323:  "EMsg_ClientUCMUnsubscribePublishedFile",
	7324:  "EMsg_ClientUCMUnsubscribePublishedFileResponse",
	7325:  "EMsg_ClientUCMUpdatePublishedFile",
	7326:  "EMsg_ClientUCMUpdatePublishedFileResponse",
	7327:  "EMsg_UCMUpdatePublishedFile",
	7328:  "EMsg_UCMUpdatePublishedFileResponse",
	7329:  "EMsg_UCMDeletePublishedFile",
	7330:  "EMsg_UCMDeletePublishedFileResponse",
	7331:  "EMsg_UCMUpdatePublishedFileStat",
	7332:  "EMsg_UCMUpdatePublishedFileBan",
	7333:  "EMsg_UCMUpdatePublishedFileBanResponse",
	7334:  "EMsg_UCMUpdateTaggedScreenshot",
	7335:  "EMsg_UCMAddTaggedScreenshot",
	7336:  "EMsg_UCMRemoveTaggedScreenshot",
	7337:  "EMsg_UCMReloadPublishedFile",
	7338:  "EMsg_UCMReloadUserFileListCaches",
	7339:  "EMsg_UCMPublishedFileReported",
	7340:  "EMsg_UCMUpdatePublishedFileIncompatibleStatus",
	7341:  "EMsg_UCMPublishedFilePreviewAdd",
	7342:  "EMsg_UCMPublishedFilePreviewAddResponse",
	7343:  "EMsg_UCMPublishedFilePreviewRemove",
	7344:  "EMsg_UCMPublishedFilePreviewRemoveResponse",
	7345:  "EMsg_UCMPublishedFilePreviewChangeSortOrder",
	7346:  "EMsg_UCMPublishedFilePreviewChangeSortOrderResponse",
	7347:  "EMsg_ClientUCMPublishedFileSubscribed",
	7348:  "EMsg_ClientUCMPublishedFileUnsubscribed",
	7349:  "EMsg_UCMPublishedFileSubscribed",
	7350:  "EMsg_UCMPublishedFileUnsubscribed",
	7351:  "EMsg_UCMPublishFile",
	7352:  "EMsg_UCMPublishFileResponse",
	7353:  "EMsg_UCMPublishedFileChildAdd",
	7354:  "EMsg_UCMPublishedFileChildAddResponse",
	7355:  "EMsg_UCMPublishedFileChildRemove",
	7356:  "EMsg_UCMPublishedFileChildRemoveResponse",
	7357:  "EMsg_UCMPublishedFileChildChangeSortOrder",
	7358:  "EMsg_UCMPublishedFileChildChangeSortOrderResponse",
	7359:  "EMsg_UCMPublishedFileParentChanged",
	7360:  "EMsg_ClientUCMGetPublishedFilesForUser",
	7361:  "EMsg_ClientUCMGetPublishedFilesForUserResponse",
	7362:  "EMsg_UCMGetPublishedFilesForUser",
	7363:  "EMsg_UCMGetPublishedFilesForUserResponse",
	7364:  "EMsg_ClientUCMSetUserPublishedFileAction",
	7365:  "EMsg_ClientUCMSetUserPublishedFileActionResponse",
	7366:  "EMsg_ClientUCMEnumeratePublishedFilesByUserAction",
	7367:  "EMsg_ClientUCMEnumeratePublishedFilesByUserActionResponse",
	7368:  "EMsg_ClientUCMPublishedFileDeleted",
	7369:  "EMsg_UCMGetUserSubscribedFiles",
	7370:  "EMsg_UCMGetUserSubscribedFilesResponse",
	7371:  "EMsg_UCMFixStatsPublishedFile",
	7372:  "EMsg_UCMDeleteOldScreenshot",
	7373:  "EMsg_UCMDeleteOldScreenshotResponse",
	7374:  "EMsg_UCMDeleteOldVideo",
	7375:  "EMsg_UCMDeleteOldVideoResponse",
	7376:  "EMsg_UCMUpdateOldScreenshotPrivacy",
	7377:  "EMsg_UCMUpdateOldScreenshotPrivacyResponse",
	7378:  "EMsg_ClientUCMEnumerateUserSubscribedFilesWithUpdates",
	7379:  "EMsg_ClientUCMEnumerateUserSubscribedFilesWithUpdatesResponse",
	7380:  "EMsg_UCMPublishedFileContentUpdated",
	7381:  "EMsg_UCMPublishedFileUpdated",
	7382:  "EMsg_ClientWorkshopItemChangesRequest",
	7383:  "EMsg_ClientWorkshopItemChangesResponse",
	7384:  "EMsg_ClientWorkshopItemInfoRequest",
	7385:  "EMsg_ClientWorkshopItemInfoResponse",
	7500:  "EMsg_FSBase",
	7501:  "EMsg_ClientRichPresenceUpload",
	7502:  "EMsg_ClientRichPresenceRequest",
	7503:  "EMsg_ClientRichPresenceInfo",
	7504:  "EMsg_FSRichPresenceRequest",
	7505:  "EMsg_FSRichPresenceResponse",
	7506:  "EMsg_FSComputeFrenematrix",
	7507:  "EMsg_FSComputeFrenematrixResponse",
	7508:  "EMsg_FSPlayStatusNotification",
	7509:  "EMsg_FSPublishPersonaStatus",
	7510:  "EMsg_FSAddOrRemoveFollower",
	7511:  "EMsg_FSAddOrRemoveFollowerResponse",
	7512:  "EMsg_FSUpdateFollowingList",
	7513:  "EMsg_FSCommentNotification",
	7514:  "EMsg_FSCommentNotificationViewed",
	7515:  "EMsg_ClientFSGetFollowerCount",
	7516:  "EMsg_ClientFSGetFollowerCountResponse",
	7517:  "EMsg_ClientFSGetIsFollowing",
	7518:  "EMsg_ClientFSGetIsFollowingResponse",
	7519:  "EMsg_ClientFSEnumerateFollowingList",
	7520:  "EMsg_ClientFSEnumerateFollowingListResponse",
	7521:  "EMsg_FSGetPendingNotificationCount",
	7522:  "EMsg_FSGetPendingNotificationCountResponse",
	7523:  "EMsg_ClientFSOfflineMessageNotification",
	7524:  "EMsg_ClientFSRequestOfflineMessageCount",
	7525:  "EMsg_ClientFSGetFriendMessageHistory",
	7526:  "EMsg_ClientFSGetFriendMessageHistoryResponse",
	7527:  "EMsg_ClientFSGetFriendMessageHistoryForOfflineMessages",
	7528:  "EMsg_ClientFSGetFriendsSteamLevels",
	7529:  "EMsg_ClientFSGetFriendsSteamLevelsResponse",
	7530:  "EMsg_FSRequestFriendData",
	7600:  "EMsg_DRMRange2",
	7601:  "EMsg_CEGVersionSetEnableDisableResponse",
	7602:  "EMsg_CEGPropStatusDRMSRequest",
	7603:  "EMsg_CEGPropStatusDRMSResponse",
	7604:  "EMsg_CEGWhackFailureReportRequest",
	7605:  "EMsg_CEGWhackFailureReportResponse",
	7606:  "EMsg_DRMSFetchVersionSet",
	7607:  "EMsg_DRMSFetchVersionSetResponse",
	7700:  "EMsg_EconBase",
	7701:  "EMsg_EconTrading_InitiateTradeRequest",
	7702:  "EMsg_EconTrading_InitiateTradeProposed",
	7703:  "EMsg_EconTrading_InitiateTradeResponse",
	7704:  "EMsg_EconTrading_InitiateTradeResult",
	7705:  "EMsg_EconTrading_StartSession",
	7706:  "EMsg_EconTrading_CancelTradeRequest",
	7707:  "EMsg_EconFlushInventoryCache",
	7708:  "EMsg_EconFlushInventoryCacheResponse",
	7711:  "EMsg_EconCDKeyProcessTransaction",
	7712:  "EMsg_EconCDKeyProcessTransactionResponse",
	7713:  "EMsg_EconGetErrorLogs",
	7714:  "EMsg_EconGetErrorLogsResponse",
	7800:  "EMsg_RMRange",
	7801:  "EMsg_RMTestVerisignOTPResponse",
	7803:  "EMsg_RMDeleteMemcachedKeys",
	7804:  "EMsg_RMRemoteInvoke",
	7805:  "EMsg_BadLoginIPList",
	7806:  "EMsg_RMMsgTraceAddTrigger",
	7807:  "EMsg_RMMsgTraceRemoveTrigger",
	7808:  "EMsg_RMMsgTraceEvent",
	7900:  "EMsg_UGSBase",
	7901:  "EMsg_ClientUGSGetGlobalStats",
	7902:  "EMsg_ClientUGSGetGlobalStatsResponse",
	8000:  "EMsg_StoreBase",
	8100:  "EMsg_UMQBase",
	8101:  "EMsg_UMQLogonResponse",
	8102:  "EMsg_UMQLogoffRequest",
	8103:  "EMsg_UMQLogoffResponse",
	8104:  "EMsg_UMQSendChatMessage",
	8105:  "EMsg_UMQIncomingChatMessage",
	8106:  "EMsg_UMQPoll",
	8107:  "EMsg_UMQPollResults",
	8108:  "EMsg_UMQ2AM_ClientMsgBatch",
	8109:  "EMsg_UMQEnqueueMobileSalePromotions",
	8110:  "EMsg_UMQEnqueueMobileAnnouncements",
	8200:  "EMsg_WorkshopBase",
	8201:  "EMsg_WorkshopAcceptTOSResponse",
	8300:  "EMsg_WebAPIBase",
	8301:  "EMsg_WebAPIValidateOAuth2TokenResponse",
	8302:  "EMsg_WebAPIInvalidateTokensForAccount",
	8303:  "EMsg_WebAPIRegisterGCInterfaces",
	8304:  "EMsg_WebAPIInvalidateOAuthClientCache",
	8305:  "EMsg_WebAPIInvalidateOAuthTokenCache",
	8306:  "EMsg_WebAPISetSecrets",
	8400:  "EMsg_BackpackBase",
	8401:  "EMsg_BackpackAddToCurrency",
	8402:  "EMsg_BackpackAddToCurrencyResponse",
	8500:  "EMsg_CREBase",
	8501:  "EMsg_CRERankByTrend",
	8502:  "EMsg_CRERankByTrendResponse",
	8503:  "EMsg_CREItemVoteSummary",
	8504:  "EMsg_CREItemVoteSummaryResponse",
	8505:  "EMsg_CRERankByVote",
	8506:  "EMsg_CRERankByVoteResponse",
	8507:  "EMsg_CREUpdateUserPublishedItemVote",
	8508:  "EMsg_CREUpdateUserPublishedItemVoteResponse",
	8509:  "EMsg_CREGetUserPublishedItemVoteDetails",
	8510:  "EMsg_CREGetUserPublishedItemVoteDetailsResponse",
	8511:  "EMsg_CREEnumeratePublishedFiles",
	8512:  "EMsg_CREEnumeratePublishedFilesResponse",
	8513:  "EMsg_CREPublishedFileVoteAdded",
	8600:  "EMsg_SecretsBase",
	8601:  "EMsg_SecretsCredentialPairResponse",
	8602:  "EMsg_SecretsRequestServerIdentity",
	8603:  "EMsg_SecretsServerIdentityResponse",
	8604:  "EMsg_SecretsUpdateServerIdentities",
	8700:  "EMsg_BoxMonitorBase",
	8701:  "EMsg_BoxMonitorReportResponse",
	8800:  "EMsg_LogsinkBase",
	8900:  "EMsg_PICSBase",
	8901:  "EMsg_ClientPICSChangesSinceRequest",
	8902:  "EMsg_ClientPICSChangesSinceResponse",
	8903:  "EMsg_ClientPICSProductInfoRequest",
	8904:  "EMsg_ClientPICSProductInfoResponse",
	8905:  "EMsg_ClientPICSAccessTokenRequest",
	8906:  "EMsg_ClientPICSAccessTokenResponse",
	9000:  "EMsg_WorkerProcess",
	9001:  "EMsg_WorkerProcessPingResponse",
	9002:  "EMsg_WorkerProcessShutdown",
	9100:  "EMsg_DRMWorkerProcess",
	9101:  "EMsg_DRMWorkerProcessDRMAndSignResponse",
	9102:  "EMsg_DRMWorkerProcessSteamworksInfoRequest",
	9103:  "EMsg_DRMWorkerProcessSteamworksInfoResponse",
	9104:  "EMsg_DRMWorkerProcessInstallDRMDLLRequest",
	9105:  "EMsg_DRMWorkerProcessInstallDRMDLLResponse",
	9106:  "EMsg_DRMWorkerProcessSecretIdStringRequest",
	9107:  "EMsg_DRMWorkerProcessSecretIdStringResponse",
	9108:  "EMsg_DRMWorkerProcessGetDRMGuidsFromFileRequest",
	9109:  "EMsg_DRMWorkerProcessGetDRMGuidsFromFileResponse",
	9110:  "EMsg_DRMWorkerProcessInstallProcessedFilesRequest",
	9111:  "EMsg_DRMWorkerProcessInstallProcessedFilesResponse",
	9112:  "EMsg_DRMWorkerProcessExamineBlobRequest",
	9113:  "EMsg_DRMWorkerProcessExamineBlobResponse",
	9114:  "EMsg_DRMWorkerProcessDescribeSecretRequest",
	9115:  "EMsg_DRMWorkerProcessDescribeSecretResponse",
	9116:  "EMsg_DRMWorkerProcessBackfillOriginalRequest",
	9117:  "EMsg_DRMWorkerProcessBackfillOriginalResponse",
	9118:  "EMsg_DRMWorkerProcessValidateDRMDLLRequest",
	9119:  "EMsg_DRMWorkerProcessValidateDRMDLLResponse",
	9120:  "EMsg_DRMWorkerProcessValidateFileRequest",
	9121:  "EMsg_DRMWorkerProcessValidateFileResponse",
	9122:  "EMsg_DRMWorkerProcessSplitAndInstallRequest",
	9123:  "EMsg_DRMWorkerProcessSplitAndInstallResponse",
	9124:  "EMsg_DRMWorkerProcessGetBlobRequest",
	9125:  "EMsg_DRMWorkerProcessGetBlobResponse",
	9126:  "EMsg_DRMWorkerProcessEvaluateCrashRequest",
	9127:  "EMsg_DRMWorkerProcessEvaluateCrashResponse",
	9128:  "EMsg_DRMWorkerProcessAnalyzeFileRequest",
	9129:  "EMsg_DRMWorkerProcessAnalyzeFileResponse",
	9130:  "EMsg_DRMWorkerProcessUnpackBlobRequest",
	9131:  "EMsg_DRMWorkerProcessUnpackBlobResponse",
	9132:  "EMsg_DRMWorkerProcessInstallAllRequest",
	9133:  "EMsg_DRMWorkerProcessInstallAllResponse",
	9200:  "EMsg_TestWorkerProcess",
	9201:  "EMsg_TestWorkerProcessLoadUnloadModuleResponse",
	9202:  "EMsg_TestWorkerProcessServiceModuleCallRequest",
	9203:  "EMsg_TestWorkerProcessServiceModuleCallResponse",
	9300:  "EMsg_QuestServerBase",
	9330:  "EMsg_ClientGetEmoticonList",
	9331:  "EMsg_ClientEmoticonList",
	9400:  "EMsg_ClientSharedLibraryBase",
	9401:  "EMsg_SLCRequestUserSessionStatus",
	9402:  "EMsg_SLCSharedLicensesLockStatus",
	9403:  "EMsg_ClientSharedLicensesLockStatus",
	9404:  "EMsg_ClientSharedLicensesStopPlaying",
	9405:  "EMsg_ClientSharedLibraryLockStatus",
	9406:  "EMsg_ClientSharedLibraryStopPlaying",
	9407:  "EMsg_SLCOwnerLibraryChanged",
	9408:  "EMsg_SLCSharedLibraryChanged",
	9500:  "EMsg_RemoteClientBase",
	9501:  "EMsg_RemoteClientAuthResponse",
	9502:  "EMsg_RemoteClientAppStatus",
	9503:  "EMsg_RemoteClientStartStream",
	9504:  "EMsg_RemoteClientStartStreamResponse",
	9505:  "EMsg_RemoteClientPing",
	9506:  "EMsg_RemoteClientPingResponse",
	9507:  "EMsg_ClientUnlockStreaming",
	9508:  "EMsg_ClientUnlockStreamingResponse",
	9509:  "EMsg_RemoteClientAcceptEULA",
	9510:  "EMsg_RemoteClientGetControllerConfig",
	9511:  "EMsg_RemoteClientGetControllerConfigResposne",
	9512:  "EMsg_RemoteClientStreamingEnabled",
	9513:  "EMsg_ClientUnlockHEVC",
	9514:  "EMsg_ClientUnlockHEVCResponse",
	9515:  "EMsg_RemoteClientStatusRequest",
	9516:  "EMsg_RemoteClientStatusResponse",
	9600:  "EMsg_ClientConcurrentSessionsBase",
	9601:  "EMsg_ClientKickPlayingSession",
	9700:  "EMsg_ClientBroadcastBase",
	9701:  "EMsg_ClientBroadcastFrames",
	9702:  "EMsg_ClientBroadcastDisconnect",
	9703:  "EMsg_ClientBroadcastScreenshot",
	9704:  "EMsg_ClientBroadcastUploadConfig",
	9800:  "EMsg_BaseClient3",
	9801:  "EMsg_ClientVoiceCallPreAuthorizeResponse",
	9802:  "EMsg_ClientServerTimestampRequest",
	9803:  "EMsg_ClientServerTimestampResponse",
	9900:  "EMsg_ClientLANP2PBase",
	9901:  "EMsg_ClientLANP2PRequestChunkResponse",
	9999:  "EMsg_ClientLANP2PMax",
	10000: "EMsg_BaseWatchdogServer",
	10100: "EMsg_ClientSiteLicenseBase",
	10101: "EMsg_ClientSiteLicenseCheckout",
	10102: "EMsg_ClientSiteLicenseCheckoutResponse",
	10103: "EMsg_ClientSiteLicenseGetAvailableSeats",
	10104: "EMsg_ClientSiteLicenseGetAvailableSeatsResponse",
	10105: "EMsg_ClientSiteLicenseGetContentCacheInfo",
	10106: "EMsg_ClientSiteLicenseGetContentCacheInfoResponse",
	12000: "EMsg_BaseChatServer",
	12001: "EMsg_ChatServerGetPendingNotificationCountResponse",
	12100: "EMsg_BaseSecretServer",
}

func (e EMsg) String() string {
	if s, ok := EMsg_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EMsg_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EResult int32

const (
	EResult_Invalid                                  EResult = 0
	EResult_OK                                       EResult = 1
	EResult_Fail                                     EResult = 2
	EResult_NoConnection                             EResult = 3
	EResult_InvalidPassword                          EResult = 5
	EResult_LoggedInElsewhere                        EResult = 6
	EResult_InvalidProtocolVer                       EResult = 7
	EResult_InvalidParam                             EResult = 8
	EResult_FileNotFound                             EResult = 9
	EResult_Busy                                     EResult = 10
	EResult_InvalidState                             EResult = 11
	EResult_InvalidName                              EResult = 12
	EResult_InvalidEmail                             EResult = 13
	EResult_DuplicateName                            EResult = 14
	EResult_AccessDenied                             EResult = 15
	EResult_Timeout                                  EResult = 16
	EResult_Banned                                   EResult = 17
	EResult_AccountNotFound                          EResult = 18
	EResult_InvalidSteamID                           EResult = 19
	EResult_ServiceUnavailable                       EResult = 20
	EResult_NotLoggedOn                              EResult = 21
	EResult_Pending                                  EResult = 22
	EResult_EncryptionFailure                        EResult = 23
	EResult_InsufficientPrivilege                    EResult = 24
	EResult_LimitExceeded                            EResult = 25
	EResult_Revoked                                  EResult = 26
	EResult_Expired                                  EResult = 27
	EResult_AlreadyRedeemed                          EResult = 28
	EResult_DuplicateRequest                         EResult = 29
	EResult_AlreadyOwned                             EResult = 30
	EResult_IPNotFound                               EResult = 31
	EResult_PersistFailed                            EResult = 32
	EResult_LockingFailed                            EResult = 33
	EResult_LogonSessionReplaced                     EResult = 34
	EResult_ConnectFailed                            EResult = 35
	EResult_HandshakeFailed                          EResult = 36
	EResult_IOFailure                                EResult = 37
	EResult_RemoteDisconnect                         EResult = 38
	EResult_ShoppingCartNotFound                     EResult = 39
	EResult_Blocked                                  EResult = 40
	EResult_Ignored                                  EResult = 41
	EResult_NoMatch                                  EResult = 42
	EResult_AccountDisabled                          EResult = 43
	EResult_ServiceReadOnly                          EResult = 44
	EResult_AccountNotFeatured                       EResult = 45
	EResult_AdministratorOK                          EResult = 46
	EResult_ContentVersion                           EResult = 47
	EResult_TryAnotherCM                             EResult = 48
	EResult_PasswordRequiredToKickSession            EResult = 49
	EResult_AlreadyLoggedInElsewhere                 EResult = 50
	EResult_Suspended                                EResult = 51
	EResult_Cancelled                                EResult = 52
	EResult_DataCorruption                           EResult = 53
	EResult_DiskFull                                 EResult = 54
	EResult_RemoteCallFailed                         EResult = 55
	EResult_PasswordUnset                            EResult = 56
	EResult_ExternalAccountUnlinked                  EResult = 57
	EResult_PSNTicketInvalid                         EResult = 58
	EResult_ExternalAccountAlreadyLinked             EResult = 59
	EResult_RemoteFileConflict                       EResult = 60
	EResult_IllegalPassword                          EResult = 61
	EResult_SameAsPreviousValue                      EResult = 62
	EResult_AccountLogonDenied                       EResult = 63
	EResult_CannotUseOldPassword                     EResult = 64
	EResult_InvalidLoginAuthCode                     EResult = 65
	EResult_AccountLogonDeniedNoMail                 EResult = 66
	EResult_HardwareNotCapableOfIPT                  EResult = 67
	EResult_IPTInitError                             EResult = 68
	EResult_ParentalControlRestricted                EResult = 69
	EResult_FacebookQueryError                       EResult = 70
	EResult_ExpiredLoginAuthCode                     EResult = 71
	EResult_IPLoginRestrictionFailed                 EResult = 72
	EResult_AccountLockedDown                        EResult = 73
	EResult_AccountLogonDeniedVerifiedEmailRequired  EResult = 74
	EResult_NoMatchingURL                            EResult = 75
	EResult_BadResponse                              EResult = 76
	EResult_RequirePasswordReEntry                   EResult = 77
	EResult_ValueOutOfRange                          EResult = 78
	EResult_UnexpectedError                          EResult = 79
	EResult_Disabled                                 EResult = 80
	EResult_InvalidCEGSubmission                     EResult = 81
	EResult_RestrictedDevice                         EResult = 82
	EResult_RegionLocked                             EResult = 83
	EResult_RateLimitExceeded                        EResult = 84
	EResult_AccountLoginDeniedNeedTwoFactor          EResult = 85
	EResult_ItemDeleted                              EResult = 86
	EResult_AccountLoginDeniedThrottle               EResult = 87
	EResult_TwoFactorCodeMismatch                    EResult = 88
	EResult_TwoFactorActivationCodeMismatch          EResult = 89
	EResult_AccountAssociatedToMultiplePartners      EResult = 90
	EResult_NotModified                              EResult = 91
	EResult_NoMobileDevice                           EResult = 92
	EResult_TimeNotSynced                            EResult = 93
	EResult_SMSCodeFailed                            EResult = 94
	EResult_AccountLimitExceeded                     EResult = 95
	EResult_AccountActivityLimitExceeded             EResult = 96
	EResult_PhoneActivityLimitExceeded               EResult = 97
	EResult_RefundToWallet                           EResult = 98
	EResult_EmailSendFailure                         EResult = 99
	EResult_NotSettled                               EResult = 100
	EResult_NeedCaptcha                              EResult = 101
	EResult_GSLTDenied                               EResult = 102
	EResult_GSOwnerDenied                            EResult = 103
	EResult_InvalidItemType                          EResult = 104
	EResult_IPBanned                                 EResult = 105
	EResult_GSLTExpired                              EResult = 106
	EResult_InsufficientFunds                        EResult = 107
	EResult_TooManyPending                           EResult = 108
	EResult_NoSiteLicensesFound                      EResult = 109
	EResult_WGNetworkSendExceeded                    EResult = 110
	EResult_AccountNotFriends                        EResult = 111
	EResult_LimitedUserAccount                       EResult = 112
	EResult_CantRemoveItem                           EResult = 113
	EResult_AccountHasBeenDeleted                    EResult = 114
	EResult_AccountHasAnExistingUserCancelledLicense EResult = 115
	EResult_DeniedDueToCommunityCooldown             EResult = 116
)

var EResult_name = map[EResult]string{
	0:   "EResult_Invalid",
	1:   "EResult_OK",
	2:   "EResult_Fail",
	3:   "EResult_NoConnection",
	5:   "EResult_InvalidPassword",
	6:   "EResult_LoggedInElsewhere",
	7:   "EResult_InvalidProtocolVer",
	8:   "EResult_InvalidParam",
	9:   "EResult_FileNotFound",
	10:  "EResult_Busy",
	11:  "EResult_InvalidState",
	12:  "EResult_InvalidName",
	13:  "EResult_InvalidEmail",
	14:  "EResult_DuplicateName",
	15:  "EResult_AccessDenied",
	16:  "EResult_Timeout",
	17:  "EResult_Banned",
	18:  "EResult_AccountNotFound",
	19:  "EResult_InvalidSteamID",
	20:  "EResult_ServiceUnavailable",
	21:  "EResult_NotLoggedOn",
	22:  "EResult_Pending",
	23:  "EResult_EncryptionFailure",
	24:  "EResult_InsufficientPrivilege",
	25:  "EResult_LimitExceeded",
	26:  "EResult_Revoked",
	27:  "EResult_Expired",
	28:  "EResult_AlreadyRedeemed",
	29:  "EResult_DuplicateRequest",
	30:  "EResult_AlreadyOwned",
	31:  "EResult_IPNotFound",
	32:  "EResult_PersistFailed",
	33:  "EResult_LockingFailed",
	34:  "EResult_LogonSessionReplaced",
	35:  "EResult_ConnectFailed",
	36:  "EResult_HandshakeFailed",
	37:  "EResult_IOFailure",
	38:  "EResult_RemoteDisconnect",
	39:  "EResult_ShoppingCartNotFound",
	40:  "EResult_Blocked",
	41:  "EResult_Ignored",
	42:  "EResult_NoMatch",
	43:  "EResult_AccountDisabled",
	44:  "EResult_ServiceReadOnly",
	45:  "EResult_AccountNotFeatured",
	46:  "EResult_AdministratorOK",
	47:  "EResult_ContentVersion",
	48:  "EResult_TryAnotherCM",
	49:  "EResult_PasswordRequiredToKickSession",
	50:  "EResult_AlreadyLoggedInElsewhere",
	51:  "EResult_Suspended",
	52:  "EResult_Cancelled",
	53:  "EResult_DataCorruption",
	54:  "EResult_DiskFull",
	55:  "EResult_RemoteCallFailed",
	56:  "EResult_PasswordNotSet",
	57:  "EResult_ExternalAccountUnlinked",
	58:  "EResult_PSNTicketInvalid",
	59:  "EResult_ExternalAccountAlreadyLinked",
	60:  "EResult_RemoteFileConflict",
	61:  "EResult_IllegalPassword",
	62:  "EResult_SameAsPreviousValue",
	63:  "EResult_AccountLogonDenied",
	64:  "EResult_CannotUseOldPassword",
	65:  "EResult_InvalidLoginAuthCode",
	66:  "EResult_AccountLogonDeniedNoMailSent",
	67:  "EResult_HardwareNotCapableOfIPT",
	68:  "EResult_IPTInitError",
	69:  "EResult_ParentalControlRestricted",
	70:  "EResult_FacebookQueryError",
	71:  "EResult_ExpiredLoginAuthCode",
	72:  "EResult_IPLoginRestrictionFailed",
	73:  "EResult_AccountLocked",
	74:  "EResult_AccountLogonDeniedVerifiedEmailRequired",
	75:  "EResult_NoMatchingURL",
	76:  "EResult_BadResponse",
	77:  "EResult_RequirePasswordReEntry",
	78:  "EResult_ValueOutOfRange",
	79:  "EResult_UnexpectedError",
	80:  "EResult_Disabled",
	81:  "EResult_InvalidCEGSubmission",
	82:  "EResult_RestrictedDevice",
	83:  "EResult_RegionLocked",
	84:  "EResult_RateLimitExceeded",
	85:  "EResult_AccountLogonDeniedNeedTwoFactorCode",
	86:  "EResult_ItemOrEntryHasBeenDeleted",
	87:  "EResult_AccountLoginDeniedThrottle",
	88:  "EResult_TwoFactorCodeMismatch",
	89:  "EResult_TwoFactorActivationCodeMismatch",
	90:  "EResult_AccountAssociatedToMultiplePlayers",
	91:  "EResult_NotModified",
	92:  "EResult_NoMobileDeviceAvailable",
	93:  "EResult_TimeIsOutOfSync",
	94:  "EResult_SMSCodeFailed",
	95:  "EResult_TooManyAccountsAccessThisResource",
	96:  "EResult_AccountActivityLimitExceeded",
	97:  "EResult_PhoneActivityLimitExceeded",
	98:  "EResult_RefundToWallet",
	99:  "EResult_EmailSendFailure",
	100: "EResult_NotSettled",
	101: "EResult_NeedCaptcha",
	102: "EResult_GSLTDenied",
	103: "EResult_GSOwnerDenied",
	104: "EResult_InvalidItemType",
	105: "EResult_IPBanned",
	106: "EResult_GSLTExpired",
	107: "EResult_InsufficientFunds",
	108: "EResult_TooManyPending",
	109: "EResult_NoSiteLicensesFound",
	110: "EResult_WGNetworkSendExceeded",
	111: "EResult_AccountNotFriends",
	112: "EResult_LimitedUserAccount",
	113: "EResult_CantRemoveItem",
	114: "EResult_AccountHasBeenDeleted",
	115: "EResult_AccountHasAnExistingUserCancelledLicense",
	116: "EResult_DeniedDueToCommunityCooldown",
}

func (e EResult) String() string {
	if s, ok := EResult_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EResult_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EUniverse int32

const (
	EUniverse_Invalid  EUniverse = 0
	EUniverse_Public   EUniverse = 1
	EUniverse_Beta     EUniverse = 2
	EUniverse_Internal EUniverse = 3
	EUniverse_Dev      EUniverse = 4
)

var EUniverse_name = map[EUniverse]string{
	0: "EUniverse_Invalid",
	1: "EUniverse_Public",
	2: "EUniverse_Beta",
	3: "EUniverse_Internal",
	4: "EUniverse_Dev",
}

func (e EUniverse) String() string {
	if s, ok := EUniverse_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EUniverse_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatEntryType int32

const (
	EChatEntryType_Invalid          EChatEntryType = 0
	EChatEntryType_ChatMsg          EChatEntryType = 1
	EChatEntryType_Typing           EChatEntryType = 2
	EChatEntryType_InviteGame       EChatEntryType = 3
	EChatEntryType_LeftConversation EChatEntryType = 6
	EChatEntryType_Entered          EChatEntryType = 7
	EChatEntryType_WasKicked        EChatEntryType = 8
	EChatEntryType_WasBanned        EChatEntryType = 9
	EChatEntryType_Disconnected     EChatEntryType = 10
	EChatEntryType_HistoricalChat   EChatEntryType = 11
	EChatEntryType_Reserved1        EChatEntryType = 12
	EChatEntryType_Reserved2        EChatEntryType = 13
	EChatEntryType_LinkBlocked      EChatEntryType = 14
)

var EChatEntryType_name = map[EChatEntryType]string{
	0:  "EChatEntryType_Invalid",
	1:  "EChatEntryType_ChatMsg",
	2:  "EChatEntryType_Typing",
	3:  "EChatEntryType_InviteGame",
	4:  "EChatEntryType_Emote",
	5:  "EChatEntryType_LobbyGameStart",
	6:  "EChatEntryType_LeftConversation",
	7:  "EChatEntryType_Entered",
	8:  "EChatEntryType_WasKicked",
	9:  "EChatEntryType_WasBanned",
	10: "EChatEntryType_Disconnected",
	11: "EChatEntryType_HistoricalChat",
	12: "EChatEntryType_Reserved1",
	13: "EChatEntryType_Reserved2",
	14: "EChatEntryType_LinkBlocked",
}

func (e EChatEntryType) String() string {
	if s, ok := EChatEntryType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatEntryType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EPersonaState int32

const (
	EPersonaState_Offline        EPersonaState = 0
	EPersonaState_Online         EPersonaState = 1
	EPersonaState_Busy           EPersonaState = 2
	EPersonaState_Away           EPersonaState = 3
	EPersonaState_Snooze         EPersonaState = 4
	EPersonaState_LookingToTrade EPersonaState = 5
	EPersonaState_LookingToPlay  EPersonaState = 6
	EPersonaState_Invisible      EPersonaState = 7
)

var EPersonaState_name = map[EPersonaState]string{
	0: "EPersonaState_Offline",
	1: "EPersonaState_Online",
	2: "EPersonaState_Busy",
	3: "EPersonaState_Away",
	4: "EPersonaState_Snooze",
	5: "EPersonaState_LookingToTrade",
	6: "EPersonaState_LookingToPlay",
	7: "EPersonaState_Invisible",
}

func (e EPersonaState) String() string {
	if s, ok := EPersonaState_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EPersonaState_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EAccountType int32

const (
	EAccountType_Invalid        EAccountType = 0
	EAccountType_Individual     EAccountType = 1
	EAccountType_Multiseat      EAccountType = 2
	EAccountType_GameServer     EAccountType = 3
	EAccountType_AnonGameServer EAccountType = 4
	EAccountType_Pending        EAccountType = 5
	EAccountType_ContentServer  EAccountType = 6
	EAccountType_Clan           EAccountType = 7
	EAccountType_Chat           EAccountType = 8
	EAccountType_ConsoleUser    EAccountType = 9
	EAccountType_AnonUser       EAccountType = 10
)

var EAccountType_name = map[EAccountType]string{
	0:  "EAccountType_Invalid",
	1:  "EAccountType_Individual",
	2:  "EAccountType_Multiseat",
	3:  "EAccountType_GameServer",
	4:  "EAccountType_AnonGameServer",
	5:  "EAccountType_Pending",
	6:  "EAccountType_ContentServer",
	7:  "EAccountType_Clan",
	8:  "EAccountType_Chat",
	9:  "EAccountType_ConsoleUser",
	10: "EAccountType_AnonUser",
}

func (e EAccountType) String() string {
	if s, ok := EAccountType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EAccountType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EFriendRelationship int32

const (
	EFriendRelationship_None             EFriendRelationship = 0
	EFriendRelationship_Blocked          EFriendRelationship = 1
	EFriendRelationship_RequestRecipient EFriendRelationship = 2
	EFriendRelationship_Friend           EFriendRelationship = 3
	EFriendRelationship_RequestInitiator EFriendRelationship = 4
	EFriendRelationship_Ignored          EFriendRelationship = 5
	EFriendRelationship_IgnoredFriend    EFriendRelationship = 6
)

var EFriendRelationship_name = map[EFriendRelationship]string{
	0: "EFriendRelationship_None",
	1: "EFriendRelationship_Blocked",
	2: "EFriendRelationship_RequestRecipient",
	3: "EFriendRelationship_Friend",
	4: "EFriendRelationship_RequestInitiator",
	5: "EFriendRelationship_Ignored",
	6: "EFriendRelationship_IgnoredFriend",
	7: "EFriendRelationship_SuggestedFriend",
}

func (e EFriendRelationship) String() string {
	if s, ok := EFriendRelationship_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EFriendRelationship_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EAccountFlags int32

const (
	EAccountFlags_NormalUser                 EAccountFlags = 0
	EAccountFlags_PersonaNameSet             EAccountFlags = 1
	EAccountFlags_Unbannable                 EAccountFlags = 2
	EAccountFlags_PasswordSet                EAccountFlags = 4
	EAccountFlags_Support                    EAccountFlags = 8
	EAccountFlags_Admin                      EAccountFlags = 16
	EAccountFlags_Supervisor                 EAccountFlags = 32
	EAccountFlags_AppEditor                  EAccountFlags = 64
	EAccountFlags_HWIDSet                    EAccountFlags = 128
	EAccountFlags_PersonalQASet              EAccountFlags = 256
	EAccountFlags_VacBeta                    EAccountFlags = 512
	EAccountFlags_Debug                      EAccountFlags = 1024
	EAccountFlags_Disabled                   EAccountFlags = 2048
	EAccountFlags_LimitedUser                EAccountFlags = 4096
	EAccountFlags_LimitedUserForce           EAccountFlags = 8192
	EAccountFlags_EmailValidated             EAccountFlags = 16384
	EAccountFlags_MarketingTreatment         EAccountFlags = 32768
	EAccountFlags_OGGInviteOptOut            EAccountFlags = 65536
	EAccountFlags_ForcePasswordChange        EAccountFlags = 131072
	EAccountFlags_ForceEmailVerification     EAccountFlags = 262144
	EAccountFlags_LogonExtraSecurity         EAccountFlags = 524288
	EAccountFlags_LogonExtraSecurityDisabled EAccountFlags = 1048576
	EAccountFlags_Steam2MigrationComplete    EAccountFlags = 2097152
	EAccountFlags_NeedLogs                   EAccountFlags = 4194304
	EAccountFlags_Lockdown                   EAccountFlags = 8388608
	EAccountFlags_MasterAppEditor            EAccountFlags = 16777216
	EAccountFlags_BannedFromWebAPI           EAccountFlags = 33554432
	EAccountFlags_ClansOnlyFromFriends       EAccountFlags = 67108864
	EAccountFlags_GlobalModerator            EAccountFlags = 134217728
	EAccountFlags_ParentalSettings           EAccountFlags = 268435456
	EAccountFlags_ThirdPartySupport          EAccountFlags = 536870912
	EAccountFlags_NeedsSSANextSteamLogon     EAccountFlags = 1073741824
)

var EAccountFlags_name = map[EAccountFlags]string{
	0:          "EAccountFlags_NormalUser",
	1:          "EAccountFlags_PersonaNameSet",
	2:          "EAccountFlags_Unbannable",
	4:          "EAccountFlags_PasswordSet",
	8:          "EAccountFlags_Support",
	16:         "EAccountFlags_Admin",
	32:         "EAccountFlags_Supervisor",
	64:         "EAccountFlags_AppEditor",
	128:        "EAccountFlags_HWIDSet",
	256:        "EAccountFlags_PersonalQASet",
	512:        "EAccountFlags_VacBeta",
	1024:       "EAccountFlags_Debug",
	2048:       "EAccountFlags_Disabled",
	4096:       "EAccountFlags_LimitedUser",
	8192:       "EAccountFlags_LimitedUserForce",
	16384:      "EAccountFlags_EmailValidated",
	32768:      "EAccountFlags_MarketingTreatment",
	65536:      "EAccountFlags_OGGInviteOptOut",
	131072:     "EAccountFlags_ForcePasswordChange",
	262144:     "EAccountFlags_ForceEmailVerification",
	524288:     "EAccountFlags_LogonExtraSecurity",
	1048576:    "EAccountFlags_LogonExtraSecurityDisabled",
	2097152:    "EAccountFlags_Steam2MigrationComplete",
	4194304:    "EAccountFlags_NeedLogs",
	8388608:    "EAccountFlags_Lockdown",
	16777216:   "EAccountFlags_MasterAppEditor",
	33554432:   "EAccountFlags_BannedFromWebAPI",
	67108864:   "EAccountFlags_ClansOnlyFromFriends",
	134217728:  "EAccountFlags_GlobalModerator",
	268435456:  "EAccountFlags_ParentalSettings",
	536870912:  "EAccountFlags_ThirdPartySupport",
	1073741824: "EAccountFlags_NeedsSSANextSteamLogon",
}

func (e EAccountFlags) String() string {
	if s, ok := EAccountFlags_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EAccountFlags_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EClanPermission int32

const (
	EClanPermission_Nobody                EClanPermission = 0
	EClanPermission_Owner                 EClanPermission = 1
	EClanPermission_Officer               EClanPermission = 2
	EClanPermission_OwnerAndOfficer       EClanPermission = 3
	EClanPermission_Member                EClanPermission = 4
	EClanPermission_Moderator             EClanPermission = 8
	EClanPermission_OwnerOfficerModerator EClanPermission = EClanPermission_Owner | EClanPermission_Officer | EClanPermission_Moderator
	EClanPermission_AllMembers            EClanPermission = EClanPermission_Owner | EClanPermission_Officer | EClanPermission_Moderator | EClanPermission_Member
	EClanPermission_OGGGameOwner          EClanPermission = 16
	EClanPermission_NonMember             EClanPermission = 128
	EClanPermission_MemberAllowed         EClanPermission = EClanPermission_NonMember | EClanPermission_Member
	EClanPermission_ModeratorAllowed      EClanPermission = EClanPermission_NonMember | EClanPermission_Member | EClanPermission_Moderator
	EClanPermission_OfficerAllowed        EClanPermission = EClanPermission_NonMember | EClanPermission_Member | EClanPermission_Moderator | EClanPermission_Officer
	EClanPermission_OwnerAllowed          EClanPermission = EClanPermission_NonMember | EClanPermission_Member | EClanPermission_Moderator | EClanPermission_Officer | EClanPermission_Owner
	EClanPermission_Anybody               EClanPermission = EClanPermission_NonMember | EClanPermission_Member | EClanPermission_Moderator | EClanPermission_Officer | EClanPermission_Owner
)

var EClanPermission_name = map[EClanPermission]string{
	0:   "EClanPermission_Nobody",
	1:   "EClanPermission_Owner",
	2:   "EClanPermission_Officer",
	3:   "EClanPermission_OwnerAndOfficer",
	4:   "EClanPermission_Member",
	8:   "EClanPermission_Moderator",
	11:  "EClanPermission_OwnerOfficerModerator",
	15:  "EClanPermission_AllMembers",
	16:  "EClanPermission_OGGGameOwner",
	128: "EClanPermission_NonMember",
	132: "EClanPermission_MemberAllowed",
	140: "EClanPermission_ModeratorAllowed",
	142: "EClanPermission_OfficerAllowed",
	143: "EClanPermission_OwnerAllowed",
}

func (e EClanPermission) String() string {
	if s, ok := EClanPermission_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EClanPermission_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatPermission int32

const (
	EChatPermission_Close                    EChatPermission = 1
	EChatPermission_Invite                   EChatPermission = 2
	EChatPermission_Talk                     EChatPermission = 8
	EChatPermission_Kick                     EChatPermission = 16
	EChatPermission_Mute                     EChatPermission = 32
	EChatPermission_SetMetadata              EChatPermission = 64
	EChatPermission_ChangePermissions        EChatPermission = 128
	EChatPermission_Ban                      EChatPermission = 256
	EChatPermission_ChangeAccess             EChatPermission = 512
	EChatPermission_EveryoneNotInClanDefault EChatPermission = EChatPermission_Talk
	EChatPermission_EveryoneDefault          EChatPermission = EChatPermission_Talk | EChatPermission_Invite
	EChatPermission_MemberDefault            EChatPermission = EChatPermission_Ban | EChatPermission_Kick | EChatPermission_Talk | EChatPermission_Invite
	EChatPermission_OfficerDefault           EChatPermission = EChatPermission_Ban | EChatPermission_Kick | EChatPermission_Talk | EChatPermission_Invite
	EChatPermission_OwnerDefault             EChatPermission = EChatPermission_ChangeAccess | EChatPermission_Ban | EChatPermission_SetMetadata | EChatPermission_Mute | EChatPermission_Kick | EChatPermission_Talk | EChatPermission_Invite | EChatPermission_Close
	EChatPermission_Mask                     EChatPermission = 1019
)

var EChatPermission_name = map[EChatPermission]string{
	1:    "EChatPermission_Close",
	2:    "EChatPermission_Invite",
	8:    "EChatPermission_Talk",
	16:   "EChatPermission_Kick",
	32:   "EChatPermission_Mute",
	64:   "EChatPermission_SetMetadata",
	128:  "EChatPermission_ChangePermissions",
	256:  "EChatPermission_Ban",
	512:  "EChatPermission_ChangeAccess",
	10:   "EChatPermission_EveryoneDefault",
	282:  "EChatPermission_MemberDefault",
	891:  "EChatPermission_OwnerDefault",
	1019: "EChatPermission_Mask",
}

func (e EChatPermission) String() string {
	if s, ok := EChatPermission_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatPermission_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EFriendFlags int32

const (
	EFriendFlags_None                 EFriendFlags = 0
	EFriendFlags_Blocked              EFriendFlags = 1
	EFriendFlags_FriendshipRequested  EFriendFlags = 2
	EFriendFlags_Immediate            EFriendFlags = 4
	EFriendFlags_ClanMember           EFriendFlags = 8
	EFriendFlags_OnGameServer         EFriendFlags = 16
	EFriendFlags_RequestingFriendship EFriendFlags = 128
	EFriendFlags_RequestingInfo       EFriendFlags = 256
	EFriendFlags_Ignored              EFriendFlags = 512
	EFriendFlags_IgnoredFriend        EFriendFlags = 1024
	EFriendFlags_Suggested            EFriendFlags = 2048
	EFriendFlags_ChatMember           EFriendFlags = 4096
	EFriendFlags_FlagAll              EFriendFlags = 65535
)

var EFriendFlags_name = map[EFriendFlags]string{
	0:     "EFriendFlags_None",
	1:     "EFriendFlags_Blocked",
	2:     "EFriendFlags_FriendshipRequested",
	4:     "EFriendFlags_Immediate",
	8:     "EFriendFlags_ClanMember",
	16:    "EFriendFlags_OnGameServer",
	128:   "EFriendFlags_RequestingFriendship",
	256:   "EFriendFlags_RequestingInfo",
	512:   "EFriendFlags_Ignored",
	1024:  "EFriendFlags_IgnoredFriend",
	2048:  "EFriendFlags_Suggested",
	4096:  "EFriendFlags_ChatMember",
	65535: "EFriendFlags_FlagAll",
}

func (e EFriendFlags) String() string {
	if s, ok := EFriendFlags_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EFriendFlags_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EPersonaStateFlag int32

const (
	EPersonaStateFlag_HasRichPresence      EPersonaStateFlag = 1
	EPersonaStateFlag_InJoinableGame       EPersonaStateFlag = 2
	EPersonaStateFlag_Golden               EPersonaStateFlag = 4
	EPersonaStateFlag_RemotePlayTogether   EPersonaStateFlag = 8
	EPersonaStateFlag_ClientTypeWeb        EPersonaStateFlag = 256
	EPersonaStateFlag_ClientTypeMobile     EPersonaStateFlag = 512
	EPersonaStateFlag_ClientTypeTenfoot    EPersonaStateFlag = 1024
	EPersonaStateFlag_ClientTypeVR         EPersonaStateFlag = 2048
	EPersonaStateFlag_LaunchTypeGamepad    EPersonaStateFlag = 4096
	EPersonaStateFlag_LaunchTypeCompatTool EPersonaStateFlag = 8192
)

var EPersonaStateFlag_name = map[EPersonaStateFlag]string{
	1:    "EPersonaStateFlag_HasRichPresence",
	2:    "EPersonaStateFlag_InJoinableGame",
	4:    "EPersonaStateFlag_Golden",
	8:    "EPersonaStateFlag_RemotePlayTogether",
	256:  "EPersonaStateFlag_OnlineUsingWeb",
	512:  "EPersonaStateFlag_OnlineUsingMobile",
	1024: "EPersonaStateFlag_OnlineUsingBigPicture",
	2048: "EPersonaStateFlag_OnlineUsingVR",
	4096: "EPersonaStateFlag_LaunchTypeGamepad",
	8192: "EPersonaStateFlag_LaunchTypeCompatTool",
}

func (e EPersonaStateFlag) String() string {
	if s, ok := EPersonaStateFlag_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EPersonaStateFlag_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EClientPersonaStateFlag int32

const (
	EClientPersonaStateFlag_Status        EClientPersonaStateFlag = 1
	EClientPersonaStateFlag_PlayerName    EClientPersonaStateFlag = 2
	EClientPersonaStateFlag_QueryPort     EClientPersonaStateFlag = 4
	EClientPersonaStateFlag_SourceID      EClientPersonaStateFlag = 8
	EClientPersonaStateFlag_Presence      EClientPersonaStateFlag = 16
	EClientPersonaStateFlag_LastSeen      EClientPersonaStateFlag = 64
	EClientPersonaStateFlag_ClanInfo      EClientPersonaStateFlag = 128 // Deprecated: renamed to UserClanRank
	EClientPersonaStateFlag_UserClanRank  EClientPersonaStateFlag = 128
	EClientPersonaStateFlag_GameExtraInfo EClientPersonaStateFlag = 256
	EClientPersonaStateFlag_GameDataBlob  EClientPersonaStateFlag = 512
	EClientPersonaStateFlag_ClanTag       EClientPersonaStateFlag = 1024 // Deprecated: renamed to ClanData
	EClientPersonaStateFlag_ClanData      EClientPersonaStateFlag = 1024
	EClientPersonaStateFlag_Facebook      EClientPersonaStateFlag = 2048
	EClientPersonaStateFlag_RichPresence  EClientPersonaStateFlag = 4096
	EClientPersonaStateFlag_Broadcast     EClientPersonaStateFlag = 8192
	EClientPersonaStateFlag_Watching      EClientPersonaStateFlag = 16384
)

var EClientPersonaStateFlag_name = map[EClientPersonaStateFlag]string{
	1:     "EClientPersonaStateFlag_Status",
	2:     "EClientPersonaStateFlag_PlayerName",
	4:     "EClientPersonaStateFlag_QueryPort",
	8:     "EClientPersonaStateFlag_SourceID",
	16:    "EClientPersonaStateFlag_Presence",
	32:    "EClientPersonaStateFlag_Metadata",
	64:    "EClientPersonaStateFlag_LastSeen",
	128:   "EClientPersonaStateFlag_ClanInfo",
	256:   "EClientPersonaStateFlag_GameExtraInfo",
	512:   "EClientPersonaStateFlag_GameDataBlob",
	1024:  "EClientPersonaStateFlag_ClanTag",
	2048:  "EClientPersonaStateFlag_Facebook",
	4096:  "EClientPersonaStateFlag_RichPresence",
	8192:  "EClientPersonaStateFlag_Broadcast",
	16384: "EClientPersonaStateFlag_Watching",
}

func (e EClientPersonaStateFlag) String() string {
	if s, ok := EClientPersonaStateFlag_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EClientPersonaStateFlag_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EAppUsageEvent int32

const (
	EAppUsageEvent_GameLaunch            EAppUsageEvent = 1
	EAppUsageEvent_GameLaunchTrial       EAppUsageEvent = 2
	EAppUsageEvent_Media                 EAppUsageEvent = 3
	EAppUsageEvent_PreloadStart          EAppUsageEvent = 4
	EAppUsageEvent_PreloadFinish         EAppUsageEvent = 5
	EAppUsageEvent_MarketingMessageView  EAppUsageEvent = 6
	EAppUsageEvent_InGameAdViewed        EAppUsageEvent = 7
	EAppUsageEvent_GameLaunchFreeWeekend EAppUsageEvent = 8
)

var EAppUsageEvent_name = map[EAppUsageEvent]string{
	1: "EAppUsageEvent_GameLaunch",
	2: "EAppUsageEvent_GameLaunchTrial",
	3: "EAppUsageEvent_Media",
	4: "EAppUsageEvent_PreloadStart",
	5: "EAppUsageEvent_PreloadFinish",
	6: "EAppUsageEvent_MarketingMessageView",
	7: "EAppUsageEvent_InGameAdViewed",
	8: "EAppUsageEvent_GameLaunchFreeWeekend",
}

func (e EAppUsageEvent) String() string {
	if s, ok := EAppUsageEvent_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EAppUsageEvent_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ELicenseFlags int32

const (
	ELicenseFlags_None                         ELicenseFlags = 0
	ELicenseFlags_Renew                        ELicenseFlags = 0x01
	ELicenseFlags_RenewalFailed                ELicenseFlags = 0x02
	ELicenseFlags_Pending                      ELicenseFlags = 0x04
	ELicenseFlags_Expired                      ELicenseFlags = 0x08
	ELicenseFlags_CancelledByUser              ELicenseFlags = 0x10
	ELicenseFlags_CancelledByAdmin             ELicenseFlags = 0x20
	ELicenseFlags_LowViolenceContent           ELicenseFlags = 0x40
	ELicenseFlags_ImportedFromSteam2           ELicenseFlags = 0x80
	ELicenseFlags_ForceRunRestriction          ELicenseFlags = 0x100
	ELicenseFlags_RegionRestrictionExpired     ELicenseFlags = 0x200
	ELicenseFlags_CancelledByFriendlyFraudLock ELicenseFlags = 0x400
	ELicenseFlags_NotActivated                 ELicenseFlags = 0x800
)

var ELicenseFlags_name = map[ELicenseFlags]string{
	0:    "ELicenseFlags_None",
	1:    "ELicenseFlags_Renew",
	2:    "ELicenseFlags_RenewalFailed",
	4:    "ELicenseFlags_Pending",
	8:    "ELicenseFlags_Expired",
	16:   "ELicenseFlags_CancelledByUser",
	32:   "ELicenseFlags_CancelledByAdmin",
	64:   "ELicenseFlags_LowViolenceContent",
	128:  "ELicenseFlags_ImportedFromSteam2",
	256:  "ELicenseFlags_ForceRunRestriction",
	512:  "ELicenseFlags_RegionRestrictionExpired",
	1024: "ELicenseFlags_CancelledByFriendlyFraudLock",
	2048: "ELicenseFlags_NotActivated",
}

func (e ELicenseFlags) String() string {
	if s, ok := ELicenseFlags_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ELicenseFlags_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ELicenseType int32

const (
	ELicenseType_NoLicense                             ELicenseType = 0
	ELicenseType_SinglePurchase                        ELicenseType = 1
	ELicenseType_SinglePurchaseLimitedUse              ELicenseType = 2
	ELicenseType_RecurringCharge                       ELicenseType = 3
	ELicenseType_RecurringChargeLimitedUse             ELicenseType = 4
	ELicenseType_RecurringChargeLimitedUseWithOverages ELicenseType = 5
	ELicenseType_RecurringOption                       ELicenseType = 6
	ELicenseType_LimitedUseDelayedActivation           ELicenseType = 7
)

var ELicenseType_name = map[ELicenseType]string{
	0: "ELicenseType_NoLicense",
	1: "ELicenseType_SinglePurchase",
	2: "ELicenseType_SinglePurchaseLimitedUse",
	3: "ELicenseType_RecurringCharge",
	4: "ELicenseType_RecurringChargeLimitedUse",
	5: "ELicenseType_RecurringChargeLimitedUseWithOverages",
	6: "ELicenseType_RecurringOption",
	7: "ELicenseType_LimitedUseDelayedActivation",
}

func (e ELicenseType) String() string {
	if s, ok := ELicenseType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ELicenseType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EPaymentMethod int32

const (
	EPaymentMethod_None                   EPaymentMethod = 0
	EPaymentMethod_ActivationCode         EPaymentMethod = 1
	EPaymentMethod_CreditCard             EPaymentMethod = 2
	EPaymentMethod_Giropay                EPaymentMethod = 3
	EPaymentMethod_PayPal                 EPaymentMethod = 4
	EPaymentMethod_Ideal                  EPaymentMethod = 5
	EPaymentMethod_PaySafeCard            EPaymentMethod = 6
	EPaymentMethod_Sofort                 EPaymentMethod = 7
	EPaymentMethod_GuestPass              EPaymentMethod = 8
	EPaymentMethod_WebMoney               EPaymentMethod = 9
	EPaymentMethod_MoneyBookers           EPaymentMethod = 10
	EPaymentMethod_AliPay                 EPaymentMethod = 11
	EPaymentMethod_Yandex                 EPaymentMethod = 12
	EPaymentMethod_Kiosk                  EPaymentMethod = 13
	EPaymentMethod_Qiwi                   EPaymentMethod = 14
	EPaymentMethod_GameStop               EPaymentMethod = 15
	EPaymentMethod_HardwarePromo          EPaymentMethod = 16
	EPaymentMethod_MoPay                  EPaymentMethod = 17
	EPaymentMethod_BoletoBancario         EPaymentMethod = 18
	EPaymentMethod_BoaCompraGold          EPaymentMethod = 19
	EPaymentMethod_BancoDoBrasilOnline    EPaymentMethod = 20
	EPaymentMethod_ItauOnline             EPaymentMethod = 21
	EPaymentMethod_BradescoOnline         EPaymentMethod = 22
	EPaymentMethod_Pagseguro              EPaymentMethod = 23
	EPaymentMethod_VisaBrazil             EPaymentMethod = 24
	EPaymentMethod_AmexBrazil             EPaymentMethod = 25
	EPaymentMethod_Aura                   EPaymentMethod = 26
	EPaymentMethod_Hipercard              EPaymentMethod = 27
	EPaymentMethod_MastercardBrazil       EPaymentMethod = 28
	EPaymentMethod_DinersCardBrazil       EPaymentMethod = 29
	EPaymentMethod_AuthorizedDevice       EPaymentMethod = 30
	EPaymentMethod_MOLPoints              EPaymentMethod = 31
	EPaymentMethod_ClickAndBuy            EPaymentMethod = 32
	EPaymentMethod_Beeline                EPaymentMethod = 33
	EPaymentMethod_Konbini                EPaymentMethod = 34
	EPaymentMethod_EClubPoints            EPaymentMethod = 35
	EPaymentMethod_CreditCardJapan        EPaymentMethod = 36
	EPaymentMethod_BankTransferJapan      EPaymentMethod = 37
	EPaymentMethod_PayEasy                EPaymentMethod = 38
	EPaymentMethod_Zong                   EPaymentMethod = 39
	EPaymentMethod_CultureVoucher         EPaymentMethod = 40
	EPaymentMethod_BookVoucher            EPaymentMethod = 41
	EPaymentMethod_HappymoneyVoucher      EPaymentMethod = 42
	EPaymentMethod_ConvenientStoreVoucher EPaymentMethod = 43
	EPaymentMethod_GameVoucher            EPaymentMethod = 44
	EPaymentMethod_Multibanco             EPaymentMethod = 45
	EPaymentMethod_Payshop                EPaymentMethod = 46
	EPaymentMethod_MaestroBoaCompra       EPaymentMethod = 47
	EPaymentMethod_OXXO                   EPaymentMethod = 48
	EPaymentMethod_ToditoCash             EPaymentMethod = 49
	EPaymentMethod_Carnet                 EPaymentMethod = 50
	EPaymentMethod_SPEI                   EPaymentMethod = 51
	EPaymentMethod_ThreePay               EPaymentMethod = 52
	EPaymentMethod_IsBank                 EPaymentMethod = 53
	EPaymentMethod_Garanti                EPaymentMethod = 54
	EPaymentMethod_Akbank                 EPaymentMethod = 55
	EPaymentMethod_YapiKredi              EPaymentMethod = 56
	EPaymentMethod_Halkbank               EPaymentMethod = 57
	EPaymentMethod_BankAsya               EPaymentMethod = 58
	EPaymentMethod_Finansbank             EPaymentMethod = 59
	EPaymentMethod_DenizBank              EPaymentMethod = 60
	EPaymentMethod_PTT                    EPaymentMethod = 61
	EPaymentMethod_CashU                  EPaymentMethod = 62
	EPaymentMethod_AutoGrant              EPaymentMethod = 64
	EPaymentMethod_WebMoneyJapan          EPaymentMethod = 65
	EPaymentMethod_OneCard                EPaymentMethod = 66
	EPaymentMethod_PSE                    EPaymentMethod = 67
	EPaymentMethod_Exito                  EPaymentMethod = 68
	EPaymentMethod_Efecty                 EPaymentMethod = 69
	EPaymentMethod_Paloto                 EPaymentMethod = 70
	EPaymentMethod_PinValidda             EPaymentMethod = 71
	EPaymentMethod_MangirKart             EPaymentMethod = 72
	EPaymentMethod_BancoCreditoDePeru     EPaymentMethod = 73
	EPaymentMethod_BBVAContinental        EPaymentMethod = 74
	EPaymentMethod_SafetyPay              EPaymentMethod = 75
	EPaymentMethod_PagoEfectivo           EPaymentMethod = 76
	EPaymentMethod_Trustly                EPaymentMethod = 77
	EPaymentMethod_UnionPay               EPaymentMethod = 78
	EPaymentMethod_BitCoin                EPaymentMethod = 79
	EPaymentMethod_Wallet                 EPaymentMethod = 128
	EPaymentMethod_Valve                  EPaymentMethod = 129
	EPaymentMethod_MasterComp             EPaymentMethod = 130
	EPaymentMethod_Promotional            EPaymentMethod = 131
	EPaymentMethod_MasterSubscription     EPaymentMethod = 134
	EPaymentMethod_OEMTicket              EPaymentMethod = 256
	EPaymentMethod_Split                  EPaymentMethod = 512
	EPaymentMethod_Complimentary          EPaymentMethod = 1024
)

var EPaymentMethod_name = map[EPaymentMethod]string{
	0:    "EPaymentMethod_None",
	1:    "EPaymentMethod_ActivationCode",
	2:    "EPaymentMethod_CreditCard",
	3:    "EPaymentMethod_Giropay",
	4:    "EPaymentMethod_PayPal",
	5:    "EPaymentMethod_Ideal",
	6:    "EPaymentMethod_PaySafeCard",
	7:    "EPaymentMethod_Sofort",
	8:    "EPaymentMethod_GuestPass",
	9:    "EPaymentMethod_WebMoney",
	10:   "EPaymentMethod_MoneyBookers",
	11:   "EPaymentMethod_AliPay",
	12:   "EPaymentMethod_Yandex",
	13:   "EPaymentMethod_Kiosk",
	14:   "EPaymentMethod_Qiwi",
	15:   "EPaymentMethod_GameStop",
	16:   "EPaymentMethod_HardwarePromo",
	17:   "EPaymentMethod_MoPay",
	18:   "EPaymentMethod_BoletoBancario",
	19:   "EPaymentMethod_BoaCompraGold",
	20:   "EPaymentMethod_BancoDoBrasilOnline",
	21:   "EPaymentMethod_ItauOnline",
	22:   "EPaymentMethod_BradescoOnline",
	23:   "EPaymentMethod_Pagseguro",
	24:   "EPaymentMethod_VisaBrazil",
	25:   "EPaymentMethod_AmexBrazil",
	26:   "EPaymentMethod_Aura",
	27:   "EPaymentMethod_Hipercard",
	28:   "EPaymentMethod_MastercardBrazil",
	29:   "EPaymentMethod_DinersCardBrazil",
	30:   "EPaymentMethod_AuthorizedDevice",
	31:   "EPaymentMethod_MOLPoints",
	32:   "EPaymentMethod_ClickAndBuy",
	33:   "EPaymentMethod_Beeline",
	34:   "EPaymentMethod_Konbini",
	35:   "EPaymentMethod_EClubPoints",
	36:   "EPaymentMethod_CreditCardJapan",
	37:   "EPaymentMethod_BankTransferJapan",
	38:   "EPaymentMethod_PayEasyJapan",
	39:   "EPaymentMethod_Zong",
	40:   "EPaymentMethod_CultureVoucher",
	41:   "EPaymentMethod_BookVoucher",
	42:   "EPaymentMethod_HappymoneyVoucher",
	43:   "EPaymentMethod_ConvenientStoreVoucher",
	44:   "EPaymentMethod_GameVoucher",
	45:   "EPaymentMethod_Multibanco",
	46:   "EPaymentMethod_Payshop",
	47:   "EPaymentMethod_Maestro",
	48:   "EPaymentMethod_OXXO",
	49:   "EPaymentMethod_ToditoCash",
	50:   "EPaymentMethod_Carnet",
	51:   "EPaymentMethod_SPEI",
	52:   "EPaymentMethod_ThreePay",
	53:   "EPaymentMethod_IsBank",
	54:   "EPaymentMethod_Garanti",
	55:   "EPaymentMethod_Akbank",
	56:   "EPaymentMethod_YapiKredi",
	57:   "EPaymentMethod_Halkbank",
	58:   "EPaymentMethod_BankAsya",
	59:   "EPaymentMethod_Finansbank",
	60:   "EPaymentMethod_DenizBank",
	61:   "EPaymentMethod_PTT",
	62:   "EPaymentMethod_CashU",
	64:   "EPaymentMethod_AutoGrant",
	65:   "EPaymentMethod_WebMoneyJapan",
	66:   "EPaymentMethod_OneCard",
	67:   "EPaymentMethod_PSE",
	68:   "EPaymentMethod_Exito",
	69:   "EPaymentMethod_Efecty",
	70:   "EPaymentMethod_Paloto",
	71:   "EPaymentMethod_PinValidda",
	72:   "EPaymentMethod_MangirKart",
	73:   "EPaymentMethod_BancoCreditoDePeru",
	74:   "EPaymentMethod_BBVAContinental",
	75:   "EPaymentMethod_SafetyPay",
	76:   "EPaymentMethod_PagoEfectivo",
	77:   "EPaymentMethod_Trustly",
	78:   "EPaymentMethod_UnionPay",
	79:   "EPaymentMethod_BitCoin",
	128:  "EPaymentMethod_Wallet",
	129:  "EPaymentMethod_Valve",
	130:  "EPaymentMethod_SteamPressMaster",
	131:  "EPaymentMethod_StorePromotion",
	134:  "EPaymentMethod_MasterSubscription",
	256:  "EPaymentMethod_OEMTicket",
	512:  "EPaymentMethod_Split",
	1024: "EPaymentMethod_Complimentary",
}

func (e EPaymentMethod) String() string {
	if s, ok := EPaymentMethod_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EPaymentMethod_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EPurchaseResultDetail int32

const (
	EPurchaseResultDetail_NoDetail                                EPurchaseResultDetail = 0
	EPurchaseResultDetail_AVSFailure                              EPurchaseResultDetail = 1
	EPurchaseResultDetail_InsufficientFunds                       EPurchaseResultDetail = 2
	EPurchaseResultDetail_ContactSupport                          EPurchaseResultDetail = 3
	EPurchaseResultDetail_Timeout                                 EPurchaseResultDetail = 4
	EPurchaseResultDetail_InvalidPackage                          EPurchaseResultDetail = 5
	EPurchaseResultDetail_InvalidPaymentMethod                    EPurchaseResultDetail = 6
	EPurchaseResultDetail_InvalidData                             EPurchaseResultDetail = 7
	EPurchaseResultDetail_OthersInProgress                        EPurchaseResultDetail = 8
	EPurchaseResultDetail_AlreadyPurchased                        EPurchaseResultDetail = 9
	EPurchaseResultDetail_WrongPrice                              EPurchaseResultDetail = 10
	EPurchaseResultDetail_FraudCheckFailed                        EPurchaseResultDetail = 11
	EPurchaseResultDetail_CancelledByUser                         EPurchaseResultDetail = 12
	EPurchaseResultDetail_RestrictedCountry                       EPurchaseResultDetail = 13
	EPurchaseResultDetail_BadActivationCode                       EPurchaseResultDetail = 14
	EPurchaseResultDetail_DuplicateActivationCode                 EPurchaseResultDetail = 15
	EPurchaseResultDetail_UseOtherPaymentMethod                   EPurchaseResultDetail = 16
	EPurchaseResultDetail_UseOtherFunctionSource                  EPurchaseResultDetail = 17
	EPurchaseResultDetail_InvalidShippingAddress                  EPurchaseResultDetail = 18
	EPurchaseResultDetail_RegionNotSupported                      EPurchaseResultDetail = 19
	EPurchaseResultDetail_AcctIsBlocked                           EPurchaseResultDetail = 20
	EPurchaseResultDetail_AcctNotVerified                         EPurchaseResultDetail = 21
	EPurchaseResultDetail_InvalidAccount                          EPurchaseResultDetail = 22
	EPurchaseResultDetail_StoreBillingCountryMismatch             EPurchaseResultDetail = 23
	EPurchaseResultDetail_DoesNotOwnRequiredApp                   EPurchaseResultDetail = 24
	EPurchaseResultDetail_CanceledByNewTransaction                EPurchaseResultDetail = 25
	EPurchaseResultDetail_ForceCanceledPending                    EPurchaseResultDetail = 26
	EPurchaseResultDetail_FailCurrencyTransProvider               EPurchaseResultDetail = 27
	EPurchaseResultDetail_FailedCyberCafe                         EPurchaseResultDetail = 28
	EPurchaseResultDetail_NeedsPreApproval                        EPurchaseResultDetail = 29
	EPurchaseResultDetail_PreApprovalDenied                       EPurchaseResultDetail = 30
	EPurchaseResultDetail_WalletCurrencyMismatch                  EPurchaseResultDetail = 31
	EPurchaseResultDetail_EmailNotValidated                       EPurchaseResultDetail = 32
	EPurchaseResultDetail_ExpiredCard                             EPurchaseResultDetail = 33
	EPurchaseResultDetail_TransactionExpired                      EPurchaseResultDetail = 34
	EPurchaseResultDetail_WouldExceedMaxWallet                    EPurchaseResultDetail = 35
	EPurchaseResultDetail_MustLoginPS3AppForPurchase              EPurchaseResultDetail = 36
	EPurchaseResultDetail_CannotShipToPOBox                       EPurchaseResultDetail = 37
	EPurchaseResultDetail_InsufficientInventory                   EPurchaseResultDetail = 38
	EPurchaseResultDetail_CannotGiftShippedGoods                  EPurchaseResultDetail = 39
	EPurchaseResultDetail_CannotShipInternationally               EPurchaseResultDetail = 40
	EPurchaseResultDetail_BillingAgreementCancelled               EPurchaseResultDetail = 41
	EPurchaseResultDetail_InvalidCoupon                           EPurchaseResultDetail = 42
	EPurchaseResultDetail_ExpiredCoupon                           EPurchaseResultDetail = 43
	EPurchaseResultDetail_AccountLocked                           EPurchaseResultDetail = 44
	EPurchaseResultDetail_OtherAbortableInProgress                EPurchaseResultDetail = 45
	EPurchaseResultDetail_ExceededSteamLimit                      EPurchaseResultDetail = 46
	EPurchaseResultDetail_OverlappingPackagesInCart               EPurchaseResultDetail = 47
	EPurchaseResultDetail_NoWallet                                EPurchaseResultDetail = 48
	EPurchaseResultDetail_NoCachedPaymentMethod                   EPurchaseResultDetail = 49
	EPurchaseResultDetail_CannotRedeemCodeFromClient              EPurchaseResultDetail = 50
	EPurchaseResultDetail_PurchaseAmountNoSupportedByProvider     EPurchaseResultDetail = 51
	EPurchaseResultDetail_OverlappingPackagesInPendingTransaction EPurchaseResultDetail = 52
	EPurchaseResultDetail_RateLimited                             EPurchaseResultDetail = 53
	EPurchaseResultDetail_OwnsExcludedApp                         EPurchaseResultDetail = 54
	EPurchaseResultDetail_CreditCardBinMismatchesType             EPurchaseResultDetail = 55
	EPurchaseResultDetail_CartValueTooHigh                        EPurchaseResultDetail = 56
	EPurchaseResultDetail_BillingAgreementAlreadyExists           EPurchaseResultDetail = 57
	EPurchaseResultDetail_POSACodeNotActivated                    EPurchaseResultDetail = 58
	EPurchaseResultDetail_CannotShipToCountry                     EPurchaseResultDetail = 59
	EPurchaseResultDetail_HungTransactionCancelled                EPurchaseResultDetail = 60
	EPurchaseResultDetail_PaypalInternalError                     EPurchaseResultDetail = 61
	EPurchaseResultDetail_UnknownGlobalCollectError               EPurchaseResultDetail = 62
	EPurchaseResultDetail_InvalidTaxAddress                       EPurchaseResultDetail = 63
	EPurchaseResultDetail_PhysicalProductLimitExceeded            EPurchaseResultDetail = 64
	EPurchaseResultDetail_PurchaseCannotBeReplayed                EPurchaseResultDetail = 65
	EPurchaseResultDetail_DelayedCompletion                       EPurchaseResultDetail = 66
	EPurchaseResultDetail_BundleTypeCannotBeGifted                EPurchaseResultDetail = 67
	EPurchaseResultDetail_BlockedByUSGov                          EPurchaseResultDetail = 68
	EPurchaseResultDetail_ItemsReservedForCommercialUse           EPurchaseResultDetail = 69
	EPurchaseResultDetail_GiftAlreadyOwned                        EPurchaseResultDetail = 70
	EPurchaseResultDetail_GiftInvalidForRecipientRegion           EPurchaseResultDetail = 71
	EPurchaseResultDetail_GiftPricingImbalance                    EPurchaseResultDetail = 72
	EPurchaseResultDetail_GiftRecipientNotSpecified               EPurchaseResultDetail = 73
	EPurchaseResultDetail_ItemsNotAllowedForCommercialUse         EPurchaseResultDetail = 74
	EPurchaseResultDetail_BusinessStoreCountryCodeMismatch        EPurchaseResultDetail = 75
	EPurchaseResultDetail_UserAssociatedWithManyCafes             EPurchaseResultDetail = 76
	EPurchaseResultDetail_UserNotAssociatedWithCafe               EPurchaseResultDetail = 77
	EPurchaseResultDetail_AddressInvalid                          EPurchaseResultDetail = 78
	EPurchaseResultDetail_CreditCardNumberInvalid                 EPurchaseResultDetail = 79
	EPurchaseResultDetail_CannotShipToMilitaryPostOffice          EPurchaseResultDetail = 80
	EPurchaseResultDetail_BillingNameInvalidResemblesCreditCard   EPurchaseResultDetail = 81
	EPurchaseResultDetail_PaymentMethodTemporarilyUnavailable     EPurchaseResultDetail = 82
	EPurchaseResultDetail_PaymentMethodNotSupportedForProduct     EPurchaseResultDetail = 83
)

var EPurchaseResultDetail_name = map[EPurchaseResultDetail]string{
	0:  "EPurchaseResultDetail_NoDetail",
	1:  "EPurchaseResultDetail_AVSFailure",
	2:  "EPurchaseResultDetail_InsufficientFunds",
	3:  "EPurchaseResultDetail_ContactSupport",
	4:  "EPurchaseResultDetail_Timeout",
	5:  "EPurchaseResultDetail_InvalidPackage",
	6:  "EPurchaseResultDetail_InvalidPaymentMethod",
	7:  "EPurchaseResultDetail_InvalidData",
	8:  "EPurchaseResultDetail_OthersInProgress",
	9:  "EPurchaseResultDetail_AlreadyPurchased",
	10: "EPurchaseResultDetail_WrongPrice",
	11: "EPurchaseResultDetail_FraudCheckFailed",
	12: "EPurchaseResultDetail_CancelledByUser",
	13: "EPurchaseResultDetail_RestrictedCountry",
	14: "EPurchaseResultDetail_BadActivationCode",
	15: "EPurchaseResultDetail_DuplicateActivationCode",
	16: "EPurchaseResultDetail_UseOtherPaymentMethod",
	17: "EPurchaseResultDetail_UseOtherFunctionSource",
	18: "EPurchaseResultDetail_InvalidShippingAddress",
	19: "EPurchaseResultDetail_RegionNotSupported",
	20: "EPurchaseResultDetail_AcctIsBlocked",
	21: "EPurchaseResultDetail_AcctNotVerified",
	22: "EPurchaseResultDetail_InvalidAccount",
	23: "EPurchaseResultDetail_StoreBillingCountryMismatch",
	24: "EPurchaseResultDetail_DoesNotOwnRequiredApp",
	25: "EPurchaseResultDetail_CanceledByNewTransaction",
	26: "EPurchaseResultDetail_ForceCanceledPending",
	27: "EPurchaseResultDetail_FailCurrencyTransProvider",
	28: "EPurchaseResultDetail_FailedCyberCafe",
	29: "EPurchaseResultDetail_NeedsPreApproval",
	30: "EPurchaseResultDetail_PreApprovalDenied",
	31: "EPurchaseResultDetail_WalletCurrencyMismatch",
	32: "EPurchaseResultDetail_EmailNotValidated",
	33: "EPurchaseResultDetail_ExpiredCard",
	34: "EPurchaseResultDetail_TransactionExpired",
	35: "EPurchaseResultDetail_WouldExceedMaxWallet",
	36: "EPurchaseResultDetail_MustLoginPS3AppForPurchase",
	37: "EPurchaseResultDetail_CannotShipToPOBox",
	38: "EPurchaseResultDetail_InsufficientInventory",
	39: "EPurchaseResultDetail_CannotGiftShippedGoods",
	40: "EPurchaseResultDetail_CannotShipInternationally",
	41: "EPurchaseResultDetail_BillingAgreementCancelled",
	42: "EPurchaseResultDetail_InvalidCoupon",
	43: "EPurchaseResultDetail_ExpiredCoupon",
	44: "EPurchaseResultDetail_AccountLocked",
	45: "EPurchaseResultDetail_OtherAbortableInProgress",
	46: "EPurchaseResultDetail_ExceededSteamLimit",
	47: "EPurchaseResultDetail_OverlappingPackagesInCart",
	48: "EPurchaseResultDetail_NoWallet",
	49: "EPurchaseResultDetail_NoCachedPaymentMethod",
	50: "EPurchaseResultDetail_CannotRedeemCodeFromClient",
	51: "EPurchaseResultDetail_PurchaseAmountNoSupportedByProvider",
	52: "EPurchaseResultDetail_OverlappingPackagesInPendingTransaction",
	53: "EPurchaseResultDetail_RateLimited",
	54: "EPurchaseResultDetail_OwnsExcludedApp",
	55: "EPurchaseResultDetail_CreditCardBinMismatchesType",
	56: "EPurchaseResultDetail_CartValueTooHigh",
	57: "EPurchaseResultDetail_BillingAgreementAlreadyExists",
	58: "EPurchaseResultDetail_POSACodeNotActivated",
	59: "EPurchaseResultDetail_CannotShipToCountry",
	60: "EPurchaseResultDetail_HungTransactionCancelled",
	61: "EPurchaseResultDetail_PaypalInternalError",
	62: "EPurchaseResultDetail_UnknownGlobalCollectError",
	63: "EPurchaseResultDetail_InvalidTaxAddress",
	64: "EPurchaseResultDetail_PhysicalProductLimitExceeded",
	65: "EPurchaseResultDetail_PurchaseCannotBeReplayed",
	66: "EPurchaseResultDetail_DelayedCompletion",
	67: "EPurchaseResultDetail_BundleTypeCannotBeGifted",
	68: "EPurchaseResultDetail_BlockedByUSGov",
	69: "EPurchaseResultDetail_ItemsReservedForCommercialUse",
	70: "EPurchaseResultDetail_GiftAlreadyOwned",
	71: "EPurchaseResultDetail_GiftInvalidForRecipientRegion",
	72: "EPurchaseResultDetail_GiftPricingImbalance",
	73: "EPurchaseResultDetail_GiftRecipientNotSpecified",
	74: "EPurchaseResultDetail_ItemsNotAllowedForCommercialUse",
	75: "EPurchaseResultDetail_BusinessStoreCountryCodeMismatch",
	76: "EPurchaseResultDetail_UserAssociatedWithManyCafes",
	77: "EPurchaseResultDetail_UserNotAssociatedWithCafe",
	78: "EPurchaseResultDetail_AddressInvalid",
	79: "EPurchaseResultDetail_CreditCardNumberInvalid",
	80: "EPurchaseResultDetail_CannotShipToMilitaryPostOffice",
	81: "EPurchaseResultDetail_BillingNameInvalidResemblesCreditCard",
	82: "EPurchaseResultDetail_PaymentMethodTemporarilyUnavailable",
	83: "EPurchaseResultDetail_PaymentMethodNotSupportedForProduct",
}

func (e EPurchaseResultDetail) String() string {
	if s, ok := EPurchaseResultDetail_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EPurchaseResultDetail_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EIntroducerRouting int32

const (
	EIntroducerRouting_P2PVoiceChat  EIntroducerRouting = 1
	EIntroducerRouting_P2PNetworking EIntroducerRouting = 2
)

var EIntroducerRouting_name = map[EIntroducerRouting]string{
	0: "EIntroducerRouting_FileShare",
	1: "EIntroducerRouting_P2PVoiceChat",
	2: "EIntroducerRouting_P2PNetworking",
}

func (e EIntroducerRouting) String() string {
	if s, ok := EIntroducerRouting_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EIntroducerRouting_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EServerFlags int32

const (
	EServerFlags_None       EServerFlags = 0
	EServerFlags_Active     EServerFlags = 1
	EServerFlags_Secure     EServerFlags = 2
	EServerFlags_Dedicated  EServerFlags = 4
	EServerFlags_Linux      EServerFlags = 8
	EServerFlags_Passworded EServerFlags = 16
	EServerFlags_Private    EServerFlags = 32
)

var EServerFlags_name = map[EServerFlags]string{
	0:  "EServerFlags_None",
	1:  "EServerFlags_Active",
	2:  "EServerFlags_Secure",
	4:  "EServerFlags_Dedicated",
	8:  "EServerFlags_Linux",
	16: "EServerFlags_Passworded",
	32: "EServerFlags_Private",
}

func (e EServerFlags) String() string {
	if s, ok := EServerFlags_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EServerFlags_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EDenyReason int32

const (
	EDenyReason_InvalidVersion          EDenyReason = 1
	EDenyReason_Generic                 EDenyReason = 2
	EDenyReason_NotLoggedOn             EDenyReason = 3
	EDenyReason_NoLicense               EDenyReason = 4
	EDenyReason_Cheater                 EDenyReason = 5
	EDenyReason_LoggedInElseWhere       EDenyReason = 6
	EDenyReason_UnknownText             EDenyReason = 7
	EDenyReason_IncompatibleAnticheat   EDenyReason = 8
	EDenyReason_MemoryCorruption        EDenyReason = 9
	EDenyReason_IncompatibleSoftware    EDenyReason = 10
	EDenyReason_SteamConnectionLost     EDenyReason = 11
	EDenyReason_SteamConnectionError    EDenyReason = 12
	EDenyReason_SteamResponseTimedOut   EDenyReason = 13
	EDenyReason_SteamValidationStalled  EDenyReason = 14
	EDenyReason_SteamOwnerLeftGuestUser EDenyReason = 15
)

var EDenyReason_name = map[EDenyReason]string{
	1:  "EDenyReason_InvalidVersion",
	2:  "EDenyReason_Generic",
	3:  "EDenyReason_NotLoggedOn",
	4:  "EDenyReason_NoLicense",
	5:  "EDenyReason_Cheater",
	6:  "EDenyReason_LoggedInElseWhere",
	7:  "EDenyReason_UnknownText",
	8:  "EDenyReason_IncompatibleAnticheat",
	9:  "EDenyReason_MemoryCorruption",
	10: "EDenyReason_IncompatibleSoftware",
	11: "EDenyReason_SteamConnectionLost",
	12: "EDenyReason_SteamConnectionError",
	13: "EDenyReason_SteamResponseTimedOut",
	14: "EDenyReason_SteamValidationStalled",
	15: "EDenyReason_SteamOwnerLeftGuestUser",
}

func (e EDenyReason) String() string {
	if s, ok := EDenyReason_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EDenyReason_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EClanRank int32

const (
	EClanRank_None      EClanRank = 0
	EClanRank_Owner     EClanRank = 1
	EClanRank_Officer   EClanRank = 2
	EClanRank_Member    EClanRank = 3
	EClanRank_Moderator EClanRank = 4
)

var EClanRank_name = map[EClanRank]string{
	0: "EClanRank_None",
	1: "EClanRank_Owner",
	2: "EClanRank_Officer",
	3: "EClanRank_Member",
	4: "EClanRank_Moderator",
}

func (e EClanRank) String() string {
	if s, ok := EClanRank_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EClanRank_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EClanRelationship int32

const (
	EClanRelationship_None             EClanRelationship = 0
	EClanRelationship_Blocked          EClanRelationship = 1
	EClanRelationship_Invited          EClanRelationship = 2
	EClanRelationship_Member           EClanRelationship = 3
	EClanRelationship_Kicked           EClanRelationship = 4
	EClanRelationship_KickAcknowledged EClanRelationship = 5
	EClanRelationship_PendingApproval  EClanRelationship = 6
	EClanRelationship_RequestDenied    EClanRelationship = 7
)

var EClanRelationship_name = map[EClanRelationship]string{
	0: "EClanRelationship_None",
	1: "EClanRelationship_Blocked",
	2: "EClanRelationship_Invited",
	3: "EClanRelationship_Member",
	4: "EClanRelationship_Kicked",
	5: "EClanRelationship_KickAcknowledged",
	6: "EClanRelationship_PendingApproval",
	7: "EClanRelationship_RequestDenied",
}

func (e EClanRelationship) String() string {
	if s, ok := EClanRelationship_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EClanRelationship_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EAuthSessionResponse int32

const (
	EAuthSessionResponse_OK                           EAuthSessionResponse = 0
	EAuthSessionResponse_UserNotConnectedToSteam      EAuthSessionResponse = 1
	EAuthSessionResponse_NoLicenseOrExpired           EAuthSessionResponse = 2
	EAuthSessionResponse_VACBanned                    EAuthSessionResponse = 3
	EAuthSessionResponse_LoggedInElseWhere            EAuthSessionResponse = 4
	EAuthSessionResponse_VACCheckTimedOut             EAuthSessionResponse = 5
	EAuthSessionResponse_AuthTicketCanceled           EAuthSessionResponse = 6
	EAuthSessionResponse_AuthTicketInvalidAlreadyUsed EAuthSessionResponse = 7
	EAuthSessionResponse_AuthTicketInvalid            EAuthSessionResponse = 8
	EAuthSessionResponse_PublisherIssuedBan           EAuthSessionResponse = 9
)

var EAuthSessionResponse_name = map[EAuthSessionResponse]string{
	0: "EAuthSessionResponse_OK",
	1: "EAuthSessionResponse_UserNotConnectedToSteam",
	2: "EAuthSessionResponse_NoLicenseOrExpired",
	3: "EAuthSessionResponse_VACBanned",
	4: "EAuthSessionResponse_LoggedInElseWhere",
	5: "EAuthSessionResponse_VACCheckTimedOut",
	6: "EAuthSessionResponse_AuthTicketCanceled",
	7: "EAuthSessionResponse_AuthTicketInvalidAlreadyUsed",
	8: "EAuthSessionResponse_AuthTicketInvalid",
	9: "EAuthSessionResponse_PublisherIssuedBan",
}

func (e EAuthSessionResponse) String() string {
	if s, ok := EAuthSessionResponse_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EAuthSessionResponse_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatRoomEnterResponse int32

const (
	EChatRoomEnterResponse_Success          EChatRoomEnterResponse = 1
	EChatRoomEnterResponse_DoesntExist      EChatRoomEnterResponse = 2
	EChatRoomEnterResponse_NotAllowed       EChatRoomEnterResponse = 3
	EChatRoomEnterResponse_Full             EChatRoomEnterResponse = 4
	EChatRoomEnterResponse_Error            EChatRoomEnterResponse = 5
	EChatRoomEnterResponse_Banned           EChatRoomEnterResponse = 6
	EChatRoomEnterResponse_Limited          EChatRoomEnterResponse = 7
	EChatRoomEnterResponse_ClanDisabled     EChatRoomEnterResponse = 8
	EChatRoomEnterResponse_CommunityBan     EChatRoomEnterResponse = 9
	EChatRoomEnterResponse_MemberBlockedYou EChatRoomEnterResponse = 10
	EChatRoomEnterResponse_YouBlockedMember EChatRoomEnterResponse = 11
)

var EChatRoomEnterResponse_name = map[EChatRoomEnterResponse]string{
	1:  "EChatRoomEnterResponse_Success",
	2:  "EChatRoomEnterResponse_DoesntExist",
	3:  "EChatRoomEnterResponse_NotAllowed",
	4:  "EChatRoomEnterResponse_Full",
	5:  "EChatRoomEnterResponse_Error",
	6:  "EChatRoomEnterResponse_Banned",
	7:  "EChatRoomEnterResponse_Limited",
	8:  "EChatRoomEnterResponse_ClanDisabled",
	9:  "EChatRoomEnterResponse_CommunityBan",
	10: "EChatRoomEnterResponse_MemberBlockedYou",
	11: "EChatRoomEnterResponse_YouBlockedMember",
	12: "EChatRoomEnterResponse_NoRankingDataLobby",
	13: "EChatRoomEnterResponse_NoRankingDataUser",
	14: "EChatRoomEnterResponse_RankOutOfRange",
}

func (e EChatRoomEnterResponse) String() string {
	if s, ok := EChatRoomEnterResponse_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatRoomEnterResponse_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatRoomType int32

const (
	EChatRoomType_Friend EChatRoomType = 1
	EChatRoomType_MUC    EChatRoomType = 2
	EChatRoomType_Lobby  EChatRoomType = 3
)

var EChatRoomType_name = map[EChatRoomType]string{
	1: "EChatRoomType_Friend",
	2: "EChatRoomType_MUC",
	3: "EChatRoomType_Lobby",
}

func (e EChatRoomType) String() string {
	if s, ok := EChatRoomType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatRoomType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatInfoType int32

const (
	EChatInfoType_StateChange       EChatInfoType = 1
	EChatInfoType_InfoUpdate        EChatInfoType = 2
	EChatInfoType_MemberLimitChange EChatInfoType = 3
)

var EChatInfoType_name = map[EChatInfoType]string{
	1: "EChatInfoType_StateChange",
	2: "EChatInfoType_InfoUpdate",
	3: "EChatInfoType_MemberLimitChange",
}

func (e EChatInfoType) String() string {
	if s, ok := EChatInfoType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatInfoType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatAction int32

const (
	EChatAction_InviteChat            EChatAction = 1
	EChatAction_Kick                  EChatAction = 2
	EChatAction_Ban                   EChatAction = 3
	EChatAction_UnBan                 EChatAction = 4
	EChatAction_StartVoiceSpeak       EChatAction = 5
	EChatAction_EndVoiceSpeak         EChatAction = 6
	EChatAction_LockChat              EChatAction = 7
	EChatAction_UnlockChat            EChatAction = 8
	EChatAction_CloseChat             EChatAction = 9
	EChatAction_SetJoinable           EChatAction = 10
	EChatAction_SetUnjoinable         EChatAction = 11
	EChatAction_SetOwner              EChatAction = 12
	EChatAction_SetInvisibleToFriends EChatAction = 13
	EChatAction_SetVisibleToFriends   EChatAction = 14
	EChatAction_SetModerated          EChatAction = 15
	EChatAction_SetUnmoderated        EChatAction = 16
)

var EChatAction_name = map[EChatAction]string{
	1:  "EChatAction_InviteChat",
	2:  "EChatAction_Kick",
	3:  "EChatAction_Ban",
	4:  "EChatAction_UnBan",
	5:  "EChatAction_StartVoiceSpeak",
	6:  "EChatAction_EndVoiceSpeak",
	7:  "EChatAction_LockChat",
	8:  "EChatAction_UnlockChat",
	9:  "EChatAction_CloseChat",
	10: "EChatAction_SetJoinable",
	11: "EChatAction_SetUnjoinable",
	12: "EChatAction_SetOwner",
	13: "EChatAction_SetInvisibleToFriends",
	14: "EChatAction_SetVisibleToFriends",
	15: "EChatAction_SetModerated",
	16: "EChatAction_SetUnmoderated",
}

func (e EChatAction) String() string {
	if s, ok := EChatAction_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatAction_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatActionResult int32

const (
	EChatActionResult_Success                EChatActionResult = 1
	EChatActionResult_Error                  EChatActionResult = 2
	EChatActionResult_NotPermitted           EChatActionResult = 3
	EChatActionResult_NotAllowedOnClanMember EChatActionResult = 4
	EChatActionResult_NotAllowedOnBannedUser EChatActionResult = 5
	EChatActionResult_NotAllowedOnChatOwner  EChatActionResult = 6
	EChatActionResult_NotAllowedOnSelf       EChatActionResult = 7
	EChatActionResult_ChatDoesntExist        EChatActionResult = 8
	EChatActionResult_ChatFull               EChatActionResult = 9
	EChatActionResult_VoiceSlotsFull         EChatActionResult = 10
)

var EChatActionResult_name = map[EChatActionResult]string{
	1:  "EChatActionResult_Success",
	2:  "EChatActionResult_Error",
	3:  "EChatActionResult_NotPermitted",
	4:  "EChatActionResult_NotAllowedOnClanMember",
	5:  "EChatActionResult_NotAllowedOnBannedUser",
	6:  "EChatActionResult_NotAllowedOnChatOwner",
	7:  "EChatActionResult_NotAllowedOnSelf",
	8:  "EChatActionResult_ChatDoesntExist",
	9:  "EChatActionResult_ChatFull",
	10: "EChatActionResult_VoiceSlotsFull",
}

func (e EChatActionResult) String() string {
	if s, ok := EChatActionResult_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatActionResult_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EAppInfoSection int32

const (
	EAppInfoSection_Unknown           EAppInfoSection = 0
	EAppInfoSection_All               EAppInfoSection = 1
	EAppInfoSection_First             EAppInfoSection = 2
	EAppInfoSection_Common            EAppInfoSection = 2
	EAppInfoSection_Extended          EAppInfoSection = 3
	EAppInfoSection_Config            EAppInfoSection = 4
	EAppInfoSection_Stats             EAppInfoSection = 5
	EAppInfoSection_Install           EAppInfoSection = 6
	EAppInfoSection_Depots            EAppInfoSection = 7
	EAppInfoSection_UFS               EAppInfoSection = 10
	EAppInfoSection_OGG               EAppInfoSection = 11
	EAppInfoSection_Policies          EAppInfoSection = 13
	EAppInfoSection_SysReqs           EAppInfoSection = 14
	EAppInfoSection_Community         EAppInfoSection = 15
	EAppInfoSection_Store             EAppInfoSection = 16
	EAppInfoSection_Localization      EAppInfoSection = 17
	EAppInfoSection_Broadcastgamedata EAppInfoSection = 18
	EAppInfoSection_Computed          EAppInfoSection = 19
	EAppInfoSection_Albummetadata     EAppInfoSection = 20
)

var EAppInfoSection_name = map[EAppInfoSection]string{
	0:  "EAppInfoSection_Unknown",
	1:  "EAppInfoSection_All",
	2:  "EAppInfoSection_First",
	3:  "EAppInfoSection_Extended",
	4:  "EAppInfoSection_Config",
	5:  "EAppInfoSection_Stats",
	6:  "EAppInfoSection_Install",
	7:  "EAppInfoSection_Depots",
	8:  "EAppInfoSection_VAC",
	9:  "EAppInfoSection_DRM",
	10: "EAppInfoSection_UFS",
	11: "EAppInfoSection_OGG",
	12: "EAppInfoSection_Items",
	13: "EAppInfoSection_Policies",
	14: "EAppInfoSection_SysReqs",
	15: "EAppInfoSection_Community",
	16: "EAppInfoSection_Store",
	17: "EAppInfoSection_Localization",
	18: "EAppInfoSection_Broadcastgamedata",
	19: "EAppInfoSection_Computed",
	20: "EAppInfoSection_Albummetadata",
}

func (e EAppInfoSection) String() string {
	if s, ok := EAppInfoSection_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EAppInfoSection_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EContentDownloadSourceType int32

const (
	EContentDownloadSourceType_Invalid    EContentDownloadSourceType = 0
	EContentDownloadSourceType_CS         EContentDownloadSourceType = 1
	EContentDownloadSourceType_CDN        EContentDownloadSourceType = 2
	EContentDownloadSourceType_LCS        EContentDownloadSourceType = 3
	EContentDownloadSourceType_ProxyCache EContentDownloadSourceType = 4
	EContentDownloadSourceType_LANPeer    EContentDownloadSourceType = 5
	EContentDownloadSourceType_SLS        EContentDownloadSourceType = 6
	EContentDownloadSourceType_SteamCache EContentDownloadSourceType = 7
	EContentDownloadSourceType_OpenCache  EContentDownloadSourceType = 8
	EContentDownloadSourceType_LANCache   EContentDownloadSourceType = 9
)

var EContentDownloadSourceType_name = map[EContentDownloadSourceType]string{
	0: "EContentDownloadSourceType_Invalid",
	1: "EContentDownloadSourceType_CS",
	2: "EContentDownloadSourceType_CDN",
	3: "EContentDownloadSourceType_LCS",
	4: "EContentDownloadSourceType_ProxyCache",
	5: "EContentDownloadSourceType_LANPeer",
	6: "EContentDownloadSourceType_SLS",
	7: "EContentDownloadSourceType_SteamCache",
	8: "EContentDownloadSourceType_OpenCache",
	9: "EContentDownloadSourceType_LANCache",
}

func (e EContentDownloadSourceType) String() string {
	if s, ok := EContentDownloadSourceType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EContentDownloadSourceType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EPlatformType int32

const (
	EPlatformType_Unknown EPlatformType = 0
	EPlatformType_Win32   EPlatformType = 1
	EPlatformType_Win64   EPlatformType = 2
	EPlatformType_Linux64 EPlatformType = 3
	EPlatformType_OSX     EPlatformType = 4
	EPlatformType_PS3     EPlatformType = 5
	EPlatformType_Linux32 EPlatformType = 6
)

var EPlatformType_name = map[EPlatformType]string{
	0: "EPlatformType_Unknown",
	1: "EPlatformType_Win32",
	2: "EPlatformType_Win64",
	3: "EPlatformType_Linux",
	4: "EPlatformType_OSX",
	5: "EPlatformType_PS3",
	6: "EPlatformType_Linux32",
}

func (e EPlatformType) String() string {
	if s, ok := EPlatformType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EPlatformType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EOSType int32

const (
	EOSType_Unknown        EOSType = -1
	EOSType_Web            EOSType = -700
	EOSType_IOSUnknown     EOSType = -600
	EOSType_IOS1           EOSType = -599
	EOSType_IOS2           EOSType = -598
	EOSType_IOS3           EOSType = -597
	EOSType_IOS4           EOSType = -596
	EOSType_IOS5           EOSType = -595
	EOSType_IOS6           EOSType = -594
	EOSType_IOS6_1         EOSType = -593
	EOSType_IOS7           EOSType = -592
	EOSType_IOS7_1         EOSType = -591
	EOSType_IOS8           EOSType = -590
	EOSType_IOS8_1         EOSType = -589
	EOSType_IOS8_2         EOSType = -588
	EOSType_IOS8_3         EOSType = -587
	EOSType_IOS8_4         EOSType = -586
	EOSType_IOS9           EOSType = -585
	EOSType_IOS9_1         EOSType = -584
	EOSType_IOS9_2         EOSType = -583
	EOSType_IOS9_3         EOSType = -582
	EOSType_IOS10          EOSType = -581
	EOSType_IOS10_1        EOSType = -580
	EOSType_IOS10_2        EOSType = -579
	EOSType_IOS10_3        EOSType = -578
	EOSType_IOS11          EOSType = -577
	EOSType_IOS11_1        EOSType = -576
	EOSType_IOS11_2        EOSType = -575
	EOSType_IOS11_3        EOSType = -574
	EOSType_IOS11_4        EOSType = -573
	EOSType_IOS12          EOSType = -572
	EOSType_IOS12_1        EOSType = -571
	EOSType_AndroidUnknown EOSType = -500
	EOSType_Android6       EOSType = -499
	EOSType_Android7       EOSType = -498
	EOSType_Android8       EOSType = -497
	EOSType_Android9       EOSType = -496
	EOSType_UMQ            EOSType = -400
	EOSType_PS3            EOSType = -300
	EOSType_MacOSUnknown   EOSType = -102
	EOSType_MacOS104       EOSType = -101
	EOSType_MacOS105       EOSType = -100
	EOSType_MacOS1058      EOSType = -99
	EOSType_MacOS106       EOSType = -95
	EOSType_MacOS1063      EOSType = -94
	EOSType_MacOS1064_slgu EOSType = -93
	EOSType_MacOS1067      EOSType = -92
	EOSType_MacOS107       EOSType = -90
	EOSType_MacOS108       EOSType = -89
	EOSType_MacOS109       EOSType = -88
	EOSType_MacOS1010      EOSType = -87
	EOSType_MacOS1011      EOSType = -86
	EOSType_MacOS1012      EOSType = -85
	EOSType_Macos1013      EOSType = -84
	EOSType_Macos1014      EOSType = -83
	EOSType_Macos1015      EOSType = -82
	EOSType_MacOSMax       EOSType = -1
	EOSType_LinuxUnknown   EOSType = -203
	EOSType_Linux22        EOSType = -202
	EOSType_Linux24        EOSType = -201
	EOSType_Linux26        EOSType = -200
	EOSType_Linux32        EOSType = -199
	EOSType_Linux35        EOSType = -198
	EOSType_Linux36        EOSType = -197
	EOSType_Linux310       EOSType = -196
	EOSType_Linux316       EOSType = -195
	EOSType_Linux318       EOSType = -194
	EOSType_Linux3x        EOSType = -193
	EOSType_Linux4x        EOSType = -192
	EOSType_Linux41        EOSType = -191
	EOSType_Linux44        EOSType = -190
	EOSType_Linux49        EOSType = -189
	EOSType_Linux414       EOSType = -188
	EOSType_Linux419       EOSType = -187
	EOSType_Linux5x        EOSType = -186
	EOSType_LinuxMax       EOSType = -101
	EOSType_WinUnknown     EOSType = 0
	EOSType_Win311         EOSType = 1
	EOSType_Win95          EOSType = 2
	EOSType_Win98          EOSType = 3
	EOSType_WinME          EOSType = 4
	EOSType_WinNT          EOSType = 5
	EOSType_Win2000        EOSType = 6
	EOSType_WinXP          EOSType = 7
	EOSType_Win2003        EOSType = 8
	EOSType_WinVista       EOSType = 9
	EOSType_Windows7       EOSType = 10
	EOSType_Win2008        EOSType = 11
	EOSType_Win2012        EOSType = 12
	EOSType_Windows8       EOSType = 13
	EOSType_Windows81      EOSType = 14
	EOSType_Win2012R2      EOSType = 15
	EOSType_Windows10      EOSType = 16
	EOSType_Win2016        EOSType = 17
	EOSType_WinMAX         EOSType = 18
)

var EOSType_name = map[EOSType]string{
	-1:   "EOSType_Unknown",
	-700: "EOSType_Web",
	-600: "EOSType_IOSUnknown",
	-599: "EOSType_IOS1",
	-598: "EOSType_IOS2",
	-597: "EOSType_IOS3",
	-596: "EOSType_IOS4",
	-595: "EOSType_IOS5",
	-594: "EOSType_IOS6",
	-593: "EOSType_IOS6_1",
	-592: "EOSType_IOS7",
	-591: "EOSType_IOS7_1",
	-590: "EOSType_IOS8",
	-589: "EOSType_IOS8_1",
	-588: "EOSType_IOS8_2",
	-587: "EOSType_IOS8_3",
	-586: "EOSType_IOS8_4",
	-585: "EOSType_IOS9",
	-584: "EOSType_IOS9_1",
	-583: "EOSType_IOS9_2",
	-582: "EOSType_IOS9_3",
	-581: "EOSType_IOS10",
	-580: "EOSType_IOS10_1",
	-579: "EOSType_IOS10_2",
	-578: "EOSType_IOS10_3",
	-577: "EOSType_IOS11",
	-576: "EOSType_IOS11_1",
	-575: "EOSType_IOS11_2",
	-574: "EOSType_IOS11_3",
	-573: "EOSType_IOS11_4",
	-572: "EOSType_IOS12",
	-571: "EOSType_IOS12_1",
	-500: "EOSType_AndroidUnknown",
	-499: "EOSType_Android6",
	-498: "EOSType_Android7",
	-497: "EOSType_Android8",
	-496: "EOSType_Android9",
	-400: "EOSType_UMQ",
	-300: "EOSType_PS3",
	-102: "EOSType_MacOSUnknown",
	-101: "EOSType_MacOS104",
	-100: "EOSType_MacOS105",
	-99:  "EOSType_MacOS1058",
	-95:  "EOSType_MacOS106",
	-94:  "EOSType_MacOS1063",
	-93:  "EOSType_MacOS1064_slgu",
	-92:  "EOSType_MacOS1067",
	-90:  "EOSType_MacOS107",
	-89:  "EOSType_MacOS108",
	-88:  "EOSType_MacOS109",
	-87:  "EOSType_MacOS1010",
	-86:  "EOSType_MacOS1011",
	-85:  "EOSType_MacOS1012",
	-84:  "EOSType_Macos1013",
	-83:  "EOSType_Macos1014",
	-82:  "EOSType_Macos1015",
	-203: "EOSType_LinuxUnknown",
	-202: "EOSType_Linux22",
	-201: "EOSType_Linux24",
	-200: "EOSType_Linux26",
	-199: "EOSType_Linux32",
	-198: "EOSType_Linux35",
	-197: "EOSType_Linux36",
	-196: "EOSType_Linux310",
	-195: "EOSType_Linux316",
	-194: "EOSType_Linux318",
	-193: "EOSType_Linux3x",
	-192: "EOSType_Linux4x",
	-191: "EOSType_Linux41",
	-190: "EOSType_Linux44",
	-189: "EOSType_Linux49",
	-188: "EOSType_Linux414",
	-187: "EOSType_Linux419",
	-186: "EOSType_Linux5x",
	0:    "EOSType_WinUnknown",
	1:    "EOSType_Win311",
	2:    "EOSType_Win95",
	3:    "EOSType_Win98",
	4:    "EOSType_WinME",
	5:    "EOSType_WinNT",
	6:    "EOSType_Win200",
	7:    "EOSType_WinXP",
	8:    "EOSType_Win2003",
	9:    "EOSType_WinVista",
	10:   "EOSType_Win7",
	11:   "EOSType_Win2008",
	12:   "EOSType_Win2012",
	13:   "EOSType_Win8",
	14:   "EOSType_Win81",
	15:   "EOSType_Win2012R2",
	16:   "EOSType_Win10",
	17:   "EOSType_Win2016",
	18:   "EOSType_WinMAX",
}

func (e EOSType) String() string {
	if s, ok := EOSType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EOSType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EServerType int32

const (
	EServerType_Util                EServerType = -2
	EServerType_Client              EServerType = -3
	EServerType_CServer             EServerType = -4
	EServerType_CEconBase           EServerType = -5
	EServerType_Invalid             EServerType = -1
	EServerType_First               EServerType = 0
	EServerType_Shell               EServerType = 0
	EServerType_GM                  EServerType = 1
	EServerType_AM                  EServerType = 3
	EServerType_BS                  EServerType = 4
	EServerType_VS                  EServerType = 5
	EServerType_ATS                 EServerType = 6
	EServerType_CM                  EServerType = 7
	EServerType_FBS                 EServerType = 8
	EServerType_BoxMonitor          EServerType = 9
	EServerType_SS                  EServerType = 10
	EServerType_DRMS                EServerType = 11
	EServerType_Console             EServerType = 13
	EServerType_PICS                EServerType = 14
	EServerType_ContentStats        EServerType = 16
	EServerType_DP                  EServerType = 17
	EServerType_WG                  EServerType = 18
	EServerType_SM                  EServerType = 19
	EServerType_SLC                 EServerType = 20
	EServerType_UFS                 EServerType = 21
	EServerType_Community           EServerType = 24
	EServerType_AppInformation      EServerType = 26
	EServerType_Spare               EServerType = 27
	EServerType_FTS                 EServerType = 28
	EServerType_SiteLicense         EServerType = 29
	EServerType_PS                  EServerType = 30
	EServerType_IS                  EServerType = 31
	EServerType_CCS                 EServerType = 32
	EServerType_DFS                 EServerType = 33
	EServerType_LBS                 EServerType = 34
	EServerType_MDS                 EServerType = 35
	EServerType_CS                  EServerType = 36
	EServerType_GC                  EServerType = 37
	EServerType_NS                  EServerType = 38
	EServerType_OGS                 EServerType = 39
	EServerType_WebAPI              EServerType = 40
	EServerType_UDS                 EServerType = 41
	EServerType_MMS                 EServerType = 42
	EServerType_GMS                 EServerType = 43
	EServerType_KGS                 EServerType = 44
	EServerType_UCM                 EServerType = 45
	EServerType_RM                  EServerType = 46
	EServerType_FS                  EServerType = 47
	EServerType_Econ                EServerType = 48
	EServerType_Backpack            EServerType = 49
	EServerType_UGS                 EServerType = 50
	EServerType_StoreFeature        EServerType = 51
	EServerType_MoneyStats          EServerType = 52
	EServerType_CRE                 EServerType = 53
	EServerType_UMQ                 EServerType = 54
	EServerType_Workshop            EServerType = 55
	EServerType_BRP                 EServerType = 56
	EServerType_GCH                 EServerType = 57
	EServerType_MPAS                EServerType = 58
	EServerType_Trade               EServerType = 59
	EServerType_Secrets             EServerType = 60
	EServerType_Logsink             EServerType = 61
	EServerType_Market              EServerType = 62
	EServerType_Quest               EServerType = 63
	EServerType_WDS                 EServerType = 64
	EServerType_ACS                 EServerType = 65
	EServerType_PNP                 EServerType = 66
	EServerType_TaxForm             EServerType = 67
	EServerType_ExternalMonitor     EServerType = 68
	EServerType_Parental            EServerType = 69
	EServerType_PartnerUpload       EServerType = 70
	EServerType_Partner             EServerType = 71
	EServerType_ES                  EServerType = 72
	EServerType_DepotWebContent     EServerType = 73
	EServerType_ExternalConfig      EServerType = 74
	EServerType_GameNotifications   EServerType = 75
	EServerType_MarketRepl          EServerType = 76
	EServerType_MarketSearch        EServerType = 77
	EServerType_Localization        EServerType = 78
	EServerType_Steam2Emulator      EServerType = 79
	EServerType_PublicTest          EServerType = 80
	EServerType_SolrMgr             EServerType = 81
	EServerType_BroadcastRelay      EServerType = 82
	EServerType_BroadcastDirectory  EServerType = 83
	EServerType_VideoManager        EServerType = 84
	EServerType_TradeOffer          EServerType = 85
	EServerType_BroadcastChat       EServerType = 86
	EServerType_Phone               EServerType = 87
	EServerType_AccountScore        EServerType = 88
	EServerType_Support             EServerType = 89
	EServerType_LogRequest          EServerType = 90
	EServerType_LogWorker           EServerType = 91
	EServerType_EmailDelivery       EServerType = 92
	EServerType_InventoryManagement EServerType = 93
	EServerType_Auth                EServerType = 94
	EServerType_StoreCatalog        EServerType = 95
	EServerType_HLTVRelay           EServerType = 96
	EServerType_IDLS                EServerType = 97
	EServerType_Perf                EServerType = 98
	EServerType_ItemInventory       EServerType = 99
	EServerType_Watchdog            EServerType = 100
	EServerType_AccountHistory      EServerType = 101
	EServerType_Chat                EServerType = 102
	EServerType_Shader              EServerType = 103
	EServerType_AccountHardware     EServerType = 104
	EServerType_WebRTC              EServerType = 105
	EServerType_Giveaway            EServerType = 106
	EServerType_ChatRoom            EServerType = 107
	EServerType_VoiceChat           EServerType = 108
	EServerType_QMS                 EServerType = 109
	EServerType_Trust               EServerType = 110
	EServerType_TimeMachine         EServerType = 111
	EServerType_VACDBMaster         EServerType = 112
	EServerType_ContentServerConfig EServerType = 113
	EServerType_Minigame            EServerType = 114
	EServerType_MLTrain             EServerType = 115
	EServerType_VACTest             EServerType = 116
	EServerType_TaxService          EServerType = 117
	EServerType_MLInference         EServerType = 118
	EServerType_UGSAggregate        EServerType = 119
	EServerType_TURN                EServerType = 120
	EServerType_RemoteClient        EServerType = 121
	EServerType_BroadcastOrigin     EServerType = 122
	EServerType_BroadcastChannel    EServerType = 123
	EServerType_SteamAR             EServerType = 124
	EServerType_China               EServerType = 125
	EServerType_CrashDump           EServerType = 126
)

var EServerType_name = map[EServerType]string{
	-2:  "EServerType_Util",
	-3:  "EServerType_Client",
	-4:  "EServerType_CServer",
	-5:  "EServerType_CEconBase",
	-1:  "EServerType_Invalid",
	0:   "EServerType_First",
	1:   "EServerType_GM",
	2:   "EServerType_BUM",
	3:   "EServerType_AM",
	4:   "EServerType_BS",
	5:   "EServerType_VS",
	6:   "EServerType_ATS",
	7:   "EServerType_CM",
	8:   "EServerType_FBS",
	9:   "EServerType_FG",
	10:  "EServerType_SS",
	11:  "EServerType_DRMS",
	12:  "EServerType_HubOBSOLETE",
	13:  "EServerType_Console",
	14:  "EServerType_ASBOBSOLETE",
	15:  "EServerType_Client",
	16:  "EServerType_BootstrapOBSOLETE",
	17:  "EServerType_DP",
	18:  "EServerType_WG",
	19:  "EServerType_SM",
	20:  "EServerType_SLC",
	21:  "EServerType_UFS",
	23:  "EServerType_Util",
	24:  "EServerType_DSS",
	25:  "EServerType_P2PRelayOBSOLETE",
	26:  "EServerType_AppInformation",
	27:  "EServerType_Spare",
	28:  "EServerType_FTS",
	29:  "EServerType_EPM",
	30:  "EServerType_PS",
	31:  "EServerType_IS",
	32:  "EServerType_CCS",
	33:  "EServerType_DFS",
	34:  "EServerType_LBS",
	35:  "EServerType_MDS",
	36:  "EServerType_CS",
	37:  "EServerType_GC",
	38:  "EServerType_NS",
	39:  "EServerType_OGS",
	40:  "EServerType_WebAPI",
	41:  "EServerType_UDS",
	42:  "EServerType_MMS",
	43:  "EServerType_GMS",
	44:  "EServerType_KGS",
	45:  "EServerType_UCM",
	46:  "EServerType_RM",
	47:  "EServerType_FS",
	48:  "EServerType_Econ",
	49:  "EServerType_Backpack",
	50:  "EServerType_UGS",
	51:  "EServerType_Store",
	52:  "EServerType_MoneyStats",
	53:  "EServerType_CRE",
	54:  "EServerType_UMQ",
	55:  "EServerType_Workshop",
	56:  "EServerType_BRP",
	57:  "EServerType_GCH",
	58:  "EServerType_MPAS",
	59:  "EServerType_Trade",
	60:  "EServerType_Secrets",
	61:  "EServerType_Logsink",
	62:  "EServerType_Market",
	63:  "EServerType_Quest",
	64:  "EServerType_WDS",
	65:  "EServerType_ACS",
	66:  "EServerType_PNP",
	67:  "EServerType_TaxForm",
	68:  "EServerType_ExternalMonitor",
	69:  "EServerType_Parental",
	70:  "EServerType_PartnerUpload",
	71:  "EServerType_Partner",
	72:  "EServerType_ES",
	73:  "EServerType_DepotWebContent",
	74:  "EServerType_ExternalConfig",
	75:  "EServerType_GameNotifications",
	76:  "EServerType_MarketRepl",
	77:  "EServerType_MarketSearch",
	78:  "EServerType_Localization",
	79:  "EServerType_Steam2Emulator",
	80:  "EServerType_PublicTest",
	81:  "EServerType_SolrMgr",
	82:  "EServerType_BroadcastRelay",
	83:  "EServerType_BroadcastDirectory",
	84:  "EServerType_VideoManager",
	85:  "EServerType_TradeOffer",
	86:  "EServerType_BroadcastChat",
	87:  "EServerType_Phone",
	88:  "EServerType_AccountScore",
	89:  "EServerType_Support",
	90:  "EServerType_LogRequest",
	91:  "EServerType_LogWorker",
	92:  "EServerType_EmailDelivery",
	93:  "EServerType_InventoryManagement",
	94:  "EServerType_Auth",
	95:  "EServerType_StoreCatalog",
	96:  "EServerType_HLTVRelay",
	97:  "EServerType_IDLS",
	98:  "EServerType_Perf",
	99:  "EServerType_ItemInventory",
	100: "EServerType_Watchdog",
	101: "EServerType_AccountHistory",
	102: "EServerType_Chat",
	103: "EServerType_Shader",
	104: "EServerType_AccountHardware",
	105: "EServerType_WebRTC",
	106: "EServerType_Giveaway",
	107: "EServerType_ChatRoom",
	108: "EServerType_VoiceChat",
	109: "EServerType_QMS",
	110: "EServerType_Trust",
	111: "EServerType_TimeMachine",
	112: "EServerType_VACDBMaster",
	113: "EServerType_ContentServerConfig",
	114: "EServerType_Minigame",
	115: "EServerType_MLTrain",
	116: "EServerType_VACTest",
	117: "EServerType_TaxService",
	118: "EServerType_MLInference",
	119: "EServerType_UGSAggregate",
	120: "EServerType_TURN",
	121: "EServerType_RemoteClient",
	122: "EServerType_BroadcastOrigin",
	123: "EServerType_BroadcastChannel",
	124: "EServerType_SteamAR",
	125: "EServerType_China",
	126: "EServerType_CrashDump",
}

func (e EServerType) String() string {
	if s, ok := EServerType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EServerType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EBillingType int32

const (
	EBillingType_NoCost                 EBillingType = 0
	EBillingType_BillOnceOnly           EBillingType = 1
	EBillingType_BillMonthly            EBillingType = 2
	EBillingType_ProofOfPrepurchaseOnly EBillingType = 3
	EBillingType_GuestPass              EBillingType = 4
	EBillingType_HardwarePromo          EBillingType = 5
	EBillingType_Gift                   EBillingType = 6
	EBillingType_AutoGrant              EBillingType = 7
	EBillingType_OEMTicket              EBillingType = 8
	EBillingType_RecurringOption        EBillingType = 9
	EBillingType_BillOnceOrCDKey        EBillingType = 10
	EBillingType_Repurchaseable         EBillingType = 11
	EBillingType_FreeOnDemand           EBillingType = 12
	EBillingType_Rental                 EBillingType = 13
	EBillingType_CommercialLicense      EBillingType = 14
	EBillingType_FreeCommercialLicense  EBillingType = 15
	EBillingType_NumBillingTypes        EBillingType = 16
)

var EBillingType_name = map[EBillingType]string{
	0:  "EBillingType_NoCost",
	1:  "EBillingType_BillOnceOnly",
	2:  "EBillingType_BillMonthly",
	3:  "EBillingType_ProofOfPrepurchaseOnly",
	4:  "EBillingType_GuestPass",
	5:  "EBillingType_HardwarePromo",
	6:  "EBillingType_Gift",
	7:  "EBillingType_AutoGrant",
	8:  "EBillingType_OEMTicket",
	9:  "EBillingType_RecurringOption",
	10: "EBillingType_BillOnceOrCDKey",
	11: "EBillingType_Repurchaseable",
	12: "EBillingType_FreeOnDemand",
	13: "EBillingType_Rental",
	14: "EBillingType_CommercialLicense",
	15: "EBillingType_FreeCommercialLicense",
	16: "EBillingType_NumBillingTypes",
}

func (e EBillingType) String() string {
	if s, ok := EBillingType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EBillingType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EActivationCodeClass uint32

const (
	EActivationCodeClass_WonCDKey     EActivationCodeClass = 0
	EActivationCodeClass_ValveCDKey   EActivationCodeClass = 1
	EActivationCodeClass_Doom3CDKey   EActivationCodeClass = 2
	EActivationCodeClass_DBLookup     EActivationCodeClass = 3
	EActivationCodeClass_Steam2010Key EActivationCodeClass = 4
	EActivationCodeClass_Max          EActivationCodeClass = 5
	EActivationCodeClass_Test         EActivationCodeClass = 2147483647
	EActivationCodeClass_Invalid      EActivationCodeClass = 4294967295
)

var EActivationCodeClass_name = map[EActivationCodeClass]string{
	0:          "EActivationCodeClass_WonCDKey",
	1:          "EActivationCodeClass_ValveCDKey",
	2:          "EActivationCodeClass_Doom3CDKey",
	3:          "EActivationCodeClass_DBLookup",
	4:          "EActivationCodeClass_Steam2010Key",
	5:          "EActivationCodeClass_Max",
	2147483647: "EActivationCodeClass_Test",
	4294967295: "EActivationCodeClass_Invalid",
}

func (e EActivationCodeClass) String() string {
	if s, ok := EActivationCodeClass_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EActivationCodeClass_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatMemberStateChange int32

const (
	EChatMemberStateChange_Entered           EChatMemberStateChange = 0x01
	EChatMemberStateChange_Left              EChatMemberStateChange = 0x02
	EChatMemberStateChange_Disconnected      EChatMemberStateChange = 0x04
	EChatMemberStateChange_Kicked            EChatMemberStateChange = 0x08
	EChatMemberStateChange_Banned            EChatMemberStateChange = 0x10
	EChatMemberStateChange_VoiceSpeaking     EChatMemberStateChange = 0x1000
	EChatMemberStateChange_VoiceDoneSpeaking EChatMemberStateChange = 0x2000
)

var EChatMemberStateChange_name = map[EChatMemberStateChange]string{
	1:    "EChatMemberStateChange_Entered",
	2:    "EChatMemberStateChange_Left",
	4:    "EChatMemberStateChange_Disconnected",
	8:    "EChatMemberStateChange_Kicked",
	16:   "EChatMemberStateChange_Banned",
	4096: "EChatMemberStateChange_VoiceSpeaking",
	8192: "EChatMemberStateChange_VoiceDoneSpeaking",
}

func (e EChatMemberStateChange) String() string {
	if s, ok := EChatMemberStateChange_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatMemberStateChange_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ERegionCode uint8

const (
	ERegionCode_USEast       ERegionCode = 0x00
	ERegionCode_USWest       ERegionCode = 0x01
	ERegionCode_SouthAmerica ERegionCode = 0x02
	ERegionCode_Europe       ERegionCode = 0x03
	ERegionCode_Asia         ERegionCode = 0x04
	ERegionCode_Australia    ERegionCode = 0x05
	ERegionCode_MiddleEast   ERegionCode = 0x06
	ERegionCode_Africa       ERegionCode = 0x07
	ERegionCode_World        ERegionCode = 0xFF
)

var ERegionCode_name = map[ERegionCode]string{
	0:   "ERegionCode_USEast",
	1:   "ERegionCode_USWest",
	2:   "ERegionCode_SouthAmerica",
	3:   "ERegionCode_Europe",
	4:   "ERegionCode_Asia",
	5:   "ERegionCode_Australia",
	6:   "ERegionCode_MiddleEast",
	7:   "ERegionCode_Africa",
	255: "ERegionCode_World",
}

func (e ERegionCode) String() string {
	if s, ok := ERegionCode_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ERegionCode_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ECurrencyCode int32

const (
	ECurrencyCode_Invalid ECurrencyCode = 0
	ECurrencyCode_USD     ECurrencyCode = 1
	ECurrencyCode_GBP     ECurrencyCode = 2
	ECurrencyCode_EUR     ECurrencyCode = 3
	ECurrencyCode_CHF     ECurrencyCode = 4
	ECurrencyCode_RUB     ECurrencyCode = 5
	ECurrencyCode_PLN     ECurrencyCode = 6
	ECurrencyCode_BRL     ECurrencyCode = 7
	ECurrencyCode_JPY     ECurrencyCode = 8
	ECurrencyCode_NOK     ECurrencyCode = 9
	ECurrencyCode_IDR     ECurrencyCode = 10
	ECurrencyCode_MYR     ECurrencyCode = 11
	ECurrencyCode_PHP     ECurrencyCode = 12
	ECurrencyCode_SGD     ECurrencyCode = 13
	ECurrencyCode_THB     ECurrencyCode = 14
	ECurrencyCode_VND     ECurrencyCode = 15
	ECurrencyCode_KRW     ECurrencyCode = 16
	ECurrencyCode_TRY     ECurrencyCode = 17
	ECurrencyCode_UAH     ECurrencyCode = 18
	ECurrencyCode_MXN     ECurrencyCode = 19
	ECurrencyCode_CAD     ECurrencyCode = 20
	ECurrencyCode_AUD     ECurrencyCode = 21
	ECurrencyCode_NZD     ECurrencyCode = 22
	ECurrencyCode_CNY     ECurrencyCode = 23
	ECurrencyCode_INR     ECurrencyCode = 24
	ECurrencyCode_CLP     ECurrencyCode = 25
	ECurrencyCode_PEN     ECurrencyCode = 26
	ECurrencyCode_COP     ECurrencyCode = 27
	ECurrencyCode_ZAR     ECurrencyCode = 28
	ECurrencyCode_HKD     ECurrencyCode = 29
	ECurrencyCode_TWD     ECurrencyCode = 30
	ECurrencyCode_SAR     ECurrencyCode = 31
	ECurrencyCode_AED     ECurrencyCode = 32
	ECurrencyCode_ARS     ECurrencyCode = 34
	ECurrencyCode_ILS     ECurrencyCode = 35
	ECurrencyCode_BYN     ECurrencyCode = 36
	ECurrencyCode_KZT     ECurrencyCode = 37
	ECurrencyCode_KWD     ECurrencyCode = 38
	ECurrencyCode_QAR     ECurrencyCode = 39
	ECurrencyCode_CRC     ECurrencyCode = 40
	ECurrencyCode_UYU     ECurrencyCode = 41
)

var ECurrencyCode_name = map[ECurrencyCode]string{
	0:  "ECurrencyCode_Invalid",
	1:  "ECurrencyCode_USD",
	2:  "ECurrencyCode_GBP",
	3:  "ECurrencyCode_EUR",
	4:  "ECurrencyCode_CHF",
	5:  "ECurrencyCode_RUB",
	6:  "ECurrencyCode_PLN",
	7:  "ECurrencyCode_BRL",
	8:  "ECurrencyCode_JPY",
	9:  "ECurrencyCode_NOK",
	10: "ECurrencyCode_IDR",
	11: "ECurrencyCode_MYR",
	12: "ECurrencyCode_PHP",
	13: "ECurrencyCode_SGD",
	14: "ECurrencyCode_THB",
	15: "ECurrencyCode_VND",
	16: "ECurrencyCode_KRW",
	17: "ECurrencyCode_TRY",
	18: "ECurrencyCode_UAH",
	19: "ECurrencyCode_MXN",
	20: "ECurrencyCode_CAD",
	21: "ECurrencyCode_AUD",
	22: "ECurrencyCode_NZD",
	23: "ECurrencyCode_CNY",
	24: "ECurrencyCode_INR",
	25: "ECurrencyCode_CLP",
	26: "ECurrencyCode_PEN",
	27: "ECurrencyCode_COP",
	28: "ECurrencyCode_ZAR",
	29: "ECurrencyCode_HKD",
	30: "ECurrencyCode_TWD",
	31: "ECurrencyCode_SAR",
	32: "ECurrencyCode_AED",
	34: "ECurrencyCode_ARS",
	35: "ECurrencyCode_ILS",
	36: "ECurrencyCode_BYN",
	37: "ECurrencyCode_KZT",
	38: "ECurrencyCode_KWD",
	39: "ECurrencyCode_QAR",
	40: "ECurrencyCode_CRC",
	41: "ECurrencyCode_UYU",
}

func (e ECurrencyCode) String() string {
	if s, ok := ECurrencyCode_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ECurrencyCode_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EDepotFileFlag int32

const (
	EDepotFileFlag_UserConfig          EDepotFileFlag = 1
	EDepotFileFlag_VersionedUserConfig EDepotFileFlag = 2
	EDepotFileFlag_Encrypted           EDepotFileFlag = 4
	EDepotFileFlag_ReadOnly            EDepotFileFlag = 8
	EDepotFileFlag_Hidden              EDepotFileFlag = 16
	EDepotFileFlag_Executable          EDepotFileFlag = 32
	EDepotFileFlag_Directory           EDepotFileFlag = 64
	EDepotFileFlag_CustomExecutable    EDepotFileFlag = 128
	EDepotFileFlag_InstallScript       EDepotFileFlag = 256
	EDepotFileFlag_Symlink             EDepotFileFlag = 512
)

var EDepotFileFlag_name = map[EDepotFileFlag]string{
	1:   "EDepotFileFlag_UserConfig",
	2:   "EDepotFileFlag_VersionedUserConfig",
	4:   "EDepotFileFlag_Encrypted",
	8:   "EDepotFileFlag_ReadOnly",
	16:  "EDepotFileFlag_Hidden",
	32:  "EDepotFileFlag_Executable",
	64:  "EDepotFileFlag_Directory",
	128: "EDepotFileFlag_CustomExecutable",
	256: "EDepotFileFlag_InstallScript",
	512: "EDepotFileFlag_Symlink",
}

func (e EDepotFileFlag) String() string {
	if s, ok := EDepotFileFlag_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EDepotFileFlag_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EWorkshopEnumerationType int32

const (
	EWorkshopEnumerationType_RankedByVote            EWorkshopEnumerationType = 0
	EWorkshopEnumerationType_Recent                  EWorkshopEnumerationType = 1
	EWorkshopEnumerationType_Trending                EWorkshopEnumerationType = 2
	EWorkshopEnumerationType_FavoriteOfFriends       EWorkshopEnumerationType = 3
	EWorkshopEnumerationType_VotedByFriends          EWorkshopEnumerationType = 4
	EWorkshopEnumerationType_ContentByFriends        EWorkshopEnumerationType = 5
	EWorkshopEnumerationType_RecentFromFollowedUsers EWorkshopEnumerationType = 6
)

var EWorkshopEnumerationType_name = map[EWorkshopEnumerationType]string{
	0: "EWorkshopEnumerationType_RankedByVote",
	1: "EWorkshopEnumerationType_Recent",
	2: "EWorkshopEnumerationType_Trending",
	3: "EWorkshopEnumerationType_FavoriteOfFriends",
	4: "EWorkshopEnumerationType_VotedByFriends",
	5: "EWorkshopEnumerationType_ContentByFriends",
	6: "EWorkshopEnumerationType_RecentFromFollowedUsers",
}

func (e EWorkshopEnumerationType) String() string {
	if s, ok := EWorkshopEnumerationType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EWorkshopEnumerationType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EPublishedFileVisibility int32

const (
	EPublishedFileVisibility_Public      EPublishedFileVisibility = 0
	EPublishedFileVisibility_FriendsOnly EPublishedFileVisibility = 1
	EPublishedFileVisibility_Private     EPublishedFileVisibility = 2
)

var EPublishedFileVisibility_name = map[EPublishedFileVisibility]string{
	0: "EPublishedFileVisibility_Public",
	1: "EPublishedFileVisibility_FriendsOnly",
	2: "EPublishedFileVisibility_Private",
}

func (e EPublishedFileVisibility) String() string {
	if s, ok := EPublishedFileVisibility_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EPublishedFileVisibility_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EWorkshopFileType int32

const (
	EWorkshopFileType_First                  EWorkshopFileType = 0
	EWorkshopFileType_Community              EWorkshopFileType = 0
	EWorkshopFileType_Microtransaction       EWorkshopFileType = 1
	EWorkshopFileType_Collection             EWorkshopFileType = 2
	EWorkshopFileType_Art                    EWorkshopFileType = 3
	EWorkshopFileType_Video                  EWorkshopFileType = 4
	EWorkshopFileType_Screenshot             EWorkshopFileType = 5
	EWorkshopFileType_Game                   EWorkshopFileType = 6
	EWorkshopFileType_Software               EWorkshopFileType = 7
	EWorkshopFileType_Concept                EWorkshopFileType = 8
	EWorkshopFileType_WebGuide               EWorkshopFileType = 9
	EWorkshopFileType_IntegratedGuide        EWorkshopFileType = 10
	EWorkshopFileType_Merch                  EWorkshopFileType = 11
	EWorkshopFileType_ControllerBinding      EWorkshopFileType = 12
	EWorkshopFileType_SteamworksAccessInvite EWorkshopFileType = 13
	EWorkshopFileType_SteamVideo             EWorkshopFileType = 14
	EWorkshopFileType_GameManagedItem        EWorkshopFileType = 15
)

var EWorkshopFileType_name = map[EWorkshopFileType]string{
	0:  "EWorkshopFileType_First",
	1:  "EWorkshopFileType_Microtransaction",
	2:  "EWorkshopFileType_Collection",
	3:  "EWorkshopFileType_Art",
	4:  "EWorkshopFileType_Video",
	5:  "EWorkshopFileType_Screenshot",
	6:  "EWorkshopFileType_Game",
	7:  "EWorkshopFileType_Software",
	8:  "EWorkshopFileType_Concept",
	9:  "EWorkshopFileType_WebGuide",
	10: "EWorkshopFileType_IntegratedGuide",
	11: "EWorkshopFileType_Merch",
	12: "EWorkshopFileType_ControllerBinding",
	13: "EWorkshopFileType_SteamworksAccessInvite",
	14: "EWorkshopFileType_SteamVideo",
	15: "EWorkshopFileType_GameManagedItem",
}

func (e EWorkshopFileType) String() string {
	if s, ok := EWorkshopFileType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EWorkshopFileType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EWorkshopFileAction int32

const (
	EWorkshopFileAction_Played    EWorkshopFileAction = 0
	EWorkshopFileAction_Completed EWorkshopFileAction = 1
)

var EWorkshopFileAction_name = map[EWorkshopFileAction]string{
	0: "EWorkshopFileAction_Played",
	1: "EWorkshopFileAction_Completed",
}

func (e EWorkshopFileAction) String() string {
	if s, ok := EWorkshopFileAction_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EWorkshopFileAction_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EEconTradeResponse int32

const (
	EEconTradeResponse_Accepted                                  EEconTradeResponse = 0
	EEconTradeResponse_Declined                                  EEconTradeResponse = 1
	EEconTradeResponse_TradeBannedInitiator                      EEconTradeResponse = 2
	EEconTradeResponse_TradeBannedTarget                         EEconTradeResponse = 3
	EEconTradeResponse_TargetAlreadyTrading                      EEconTradeResponse = 4
	EEconTradeResponse_Disabled                                  EEconTradeResponse = 5
	EEconTradeResponse_NotLoggedIn                               EEconTradeResponse = 6
	EEconTradeResponse_Cancel                                    EEconTradeResponse = 7
	EEconTradeResponse_TooSoon                                   EEconTradeResponse = 8
	EEconTradeResponse_TooSoonPenalty                            EEconTradeResponse = 9
	EEconTradeResponse_ConnectionFailed                          EEconTradeResponse = 10
	EEconTradeResponse_AlreadyTrading                            EEconTradeResponse = 11
	EEconTradeResponse_AlreadyHasTradeRequest                    EEconTradeResponse = 12
	EEconTradeResponse_NoResponse                                EEconTradeResponse = 13
	EEconTradeResponse_CyberCafeInitiator                        EEconTradeResponse = 14
	EEconTradeResponse_CyberCafeTarget                           EEconTradeResponse = 15
	EEconTradeResponse_SchoolLabInitiator                        EEconTradeResponse = 16
	EEconTradeResponse_SchoolLabTarget                           EEconTradeResponse = 16
	EEconTradeResponse_InitiatorBlockedTarget                    EEconTradeResponse = 18
	EEconTradeResponse_InitiatorNeedsVerifiedEmail               EEconTradeResponse = 20
	EEconTradeResponse_InitiatorNeedsSteamGuard                  EEconTradeResponse = 21
	EEconTradeResponse_TargetAccountCannotTrade                  EEconTradeResponse = 22
	EEconTradeResponse_InitiatorSteamGuardDuration               EEconTradeResponse = 23
	EEconTradeResponse_InitiatorPasswordResetProbation           EEconTradeResponse = 24
	EEconTradeResponse_InitiatorNewDeviceCooldown                EEconTradeResponse = 25
	EEconTradeResponse_InitiatorSentInvalidCookie                EEconTradeResponse = 26
	EEconTradeResponse_NeedsEmailConfirmation                    EEconTradeResponse = 27
	EEconTradeResponse_InitiatorRecentEmailChange                EEconTradeResponse = 28
	EEconTradeResponse_NeedsMobileConfirmation                   EEconTradeResponse = 29
	EEconTradeResponse_TradingHoldForClearedTradeOffersInitiator EEconTradeResponse = 30
	EEconTradeResponse_WouldExceedMaxAssetCount                  EEconTradeResponse = 31
	EEconTradeResponse_DisabledInRegion                          EEconTradeResponse = 32
	EEconTradeResponse_DisabledInPartnerRegion                   EEconTradeResponse = 33
	EEconTradeResponse_OKToDeliver                               EEconTradeResponse = 50
)

var EEconTradeResponse_name = map[EEconTradeResponse]string{
	0:  "EEconTradeResponse_Accepted",
	1:  "EEconTradeResponse_Declined",
	2:  "EEconTradeResponse_TradeBannedInitiator",
	3:  "EEconTradeResponse_TradeBannedTarget",
	4:  "EEconTradeResponse_TargetAlreadyTrading",
	5:  "EEconTradeResponse_Disabled",
	6:  "EEconTradeResponse_NotLoggedIn",
	7:  "EEconTradeResponse_Cancel",
	8:  "EEconTradeResponse_TooSoon",
	9:  "EEconTradeResponse_TooSoonPenalty",
	10: "EEconTradeResponse_ConnectionFailed",
	11: "EEconTradeResponse_AlreadyTrading",
	12: "EEconTradeResponse_AlreadyHasTradeRequest",
	13: "EEconTradeResponse_NoResponse",
	14: "EEconTradeResponse_CyberCafeInitiator",
	15: "EEconTradeResponse_CyberCafeTarget",
	16: "EEconTradeResponse_SchoolLabInitiator",
	18: "EEconTradeResponse_InitiatorBlockedTarget",
	20: "EEconTradeResponse_InitiatorNeedsVerifiedEmail",
	21: "EEconTradeResponse_InitiatorNeedsSteamGuard",
	22: "EEconTradeResponse_TargetAccountCannotTrade",
	23: "EEconTradeResponse_InitiatorSteamGuardDuration",
	24: "EEconTradeResponse_InitiatorPasswordResetProbation",
	25: "EEconTradeResponse_InitiatorNewDeviceCooldown",
	26: "EEconTradeResponse_InitiatorSentInvalidCookie",
	27: "EEconTradeResponse_NeedsEmailConfirmation",
	28: "EEconTradeResponse_InitiatorRecentEmailChange",
	29: "EEconTradeResponse_NeedsMobileConfirmation",
	30: "EEconTradeResponse_TradingHoldForClearedTradeOffersInitiator",
	31: "EEconTradeResponse_WouldExceedMaxAssetCount",
	32: "EEconTradeResponse_DisabledInRegion",
	33: "EEconTradeResponse_DisabledInPartnerRegion",
	50: "EEconTradeResponse_OKToDeliver",
}

func (e EEconTradeResponse) String() string {
	if s, ok := EEconTradeResponse_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EEconTradeResponse_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EMarketingMessageFlags int32

const (
	EMarketingMessageFlags_None                 EMarketingMessageFlags = 0
	EMarketingMessageFlags_HighPriority         EMarketingMessageFlags = 1
	EMarketingMessageFlags_PlatformWindows      EMarketingMessageFlags = 2
	EMarketingMessageFlags_PlatformMac          EMarketingMessageFlags = 4
	EMarketingMessageFlags_PlatformLinux        EMarketingMessageFlags = 8
	EMarketingMessageFlags_PlatformRestrictions EMarketingMessageFlags = EMarketingMessageFlags_PlatformWindows | EMarketingMessageFlags_PlatformMac | EMarketingMessageFlags_PlatformLinux
)

var EMarketingMessageFlags_name = map[EMarketingMessageFlags]string{
	0:  "EMarketingMessageFlags_None",
	1:  "EMarketingMessageFlags_HighPriority",
	2:  "EMarketingMessageFlags_PlatformWindows",
	4:  "EMarketingMessageFlags_PlatformMac",
	8:  "EMarketingMessageFlags_PlatformLinux",
	14: "EMarketingMessageFlags_PlatformRestrictions",
}

func (e EMarketingMessageFlags) String() string {
	if s, ok := EMarketingMessageFlags_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EMarketingMessageFlags_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ENewsUpdateType int32

const (
	ENewsUpdateType_AppNews      ENewsUpdateType = 0
	ENewsUpdateType_SteamAds     ENewsUpdateType = 1
	ENewsUpdateType_SteamNews    ENewsUpdateType = 2
	ENewsUpdateType_CDDBUpdate   ENewsUpdateType = 3
	ENewsUpdateType_ClientUpdate ENewsUpdateType = 4
)

var ENewsUpdateType_name = map[ENewsUpdateType]string{
	0: "ENewsUpdateType_AppNews",
	1: "ENewsUpdateType_SteamAds",
	2: "ENewsUpdateType_SteamNews",
	3: "ENewsUpdateType_CDDBUpdate",
	4: "ENewsUpdateType_ClientUpdate",
}

func (e ENewsUpdateType) String() string {
	if s, ok := ENewsUpdateType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ENewsUpdateType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ESystemIMType int32

const (
	ESystemIMType_RawText                  ESystemIMType = 0
	ESystemIMType_InvalidCard              ESystemIMType = 1
	ESystemIMType_RecurringPurchaseFailed  ESystemIMType = 2
	ESystemIMType_CardWillExpire           ESystemIMType = 3
	ESystemIMType_SubscriptionExpired      ESystemIMType = 4
	ESystemIMType_GuestPassReceived        ESystemIMType = 5
	ESystemIMType_GuestPassGranted         ESystemIMType = 6
	ESystemIMType_GiftRevoked              ESystemIMType = 7
	ESystemIMType_SupportMessage           ESystemIMType = 8
	ESystemIMType_SupportMessageClearAlert ESystemIMType = 9
)

var ESystemIMType_name = map[ESystemIMType]string{
	0: "ESystemIMType_RawText",
	1: "ESystemIMType_InvalidCard",
	2: "ESystemIMType_RecurringPurchaseFailed",
	3: "ESystemIMType_CardWillExpire",
	4: "ESystemIMType_SubscriptionExpired",
	5: "ESystemIMType_GuestPassReceived",
	6: "ESystemIMType_GuestPassGranted",
	7: "ESystemIMType_GiftRevoked",
	8: "ESystemIMType_SupportMessage",
	9: "ESystemIMType_SupportMessageClearAlert",
}

func (e ESystemIMType) String() string {
	if s, ok := ESystemIMType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ESystemIMType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatFlags int32

const (
	EChatFlags_Locked             EChatFlags = 1
	EChatFlags_InvisibleToFriends EChatFlags = 2
	EChatFlags_Moderated          EChatFlags = 4
	EChatFlags_Unjoinable         EChatFlags = 8
)

var EChatFlags_name = map[EChatFlags]string{
	1: "EChatFlags_Locked",
	2: "EChatFlags_InvisibleToFriends",
	4: "EChatFlags_Moderated",
	8: "EChatFlags_Unjoinable",
}

func (e EChatFlags) String() string {
	if s, ok := EChatFlags_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatFlags_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ERemoteStoragePlatform int32

const (
	ERemoteStoragePlatform_None     ERemoteStoragePlatform = 0
	ERemoteStoragePlatform_Windows  ERemoteStoragePlatform = 1
	ERemoteStoragePlatform_OSX      ERemoteStoragePlatform = 2
	ERemoteStoragePlatform_PS3      ERemoteStoragePlatform = 4
	ERemoteStoragePlatform_Linux    ERemoteStoragePlatform = 8
	ERemoteStoragePlatform_Switch   ERemoteStoragePlatform = 16
	ERemoteStoragePlatform_Android  ERemoteStoragePlatform = 32
	ERemoteStoragePlatform_IPhoneOS ERemoteStoragePlatform = 64
	ERemoteStoragePlatform_All      ERemoteStoragePlatform = -1
)

var ERemoteStoragePlatform_name = map[ERemoteStoragePlatform]string{
	0:  "ERemoteStoragePlatform_None",
	1:  "ERemoteStoragePlatform_Windows",
	2:  "ERemoteStoragePlatform_OSX",
	4:  "ERemoteStoragePlatform_PS3",
	8:  "ERemoteStoragePlatform_Linux",
	16: "ERemoteStoragePlatform_Switch",
	32: "ERemoteStoragePlatform_Android",
	64: "ERemoteStoragePlatform_IPhoneOS",
	-1: "ERemoteStoragePlatform_All",
}

func (e ERemoteStoragePlatform) String() string {
	if s, ok := ERemoteStoragePlatform_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ERemoteStoragePlatform_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EDRMBlobDownloadType int32

const (
	EDRMBlobDownloadType_Error        EDRMBlobDownloadType = 0
	EDRMBlobDownloadType_File         EDRMBlobDownloadType = 1
	EDRMBlobDownloadType_Parts        EDRMBlobDownloadType = 2
	EDRMBlobDownloadType_Compressed   EDRMBlobDownloadType = 4
	EDRMBlobDownloadType_AllMask      EDRMBlobDownloadType = 7
	EDRMBlobDownloadType_IsJob        EDRMBlobDownloadType = 8
	EDRMBlobDownloadType_HighPriority EDRMBlobDownloadType = 16
	EDRMBlobDownloadType_AddTimestamp EDRMBlobDownloadType = 32
	EDRMBlobDownloadType_LowPriority  EDRMBlobDownloadType = 64
)

var EDRMBlobDownloadType_name = map[EDRMBlobDownloadType]string{
	0:  "EDRMBlobDownloadType_Error",
	1:  "EDRMBlobDownloadType_File",
	2:  "EDRMBlobDownloadType_Parts",
	4:  "EDRMBlobDownloadType_Compressed",
	7:  "EDRMBlobDownloadType_AllMask",
	8:  "EDRMBlobDownloadType_IsJob",
	16: "EDRMBlobDownloadType_HighPriority",
	32: "EDRMBlobDownloadType_AddTimestamp",
	64: "EDRMBlobDownloadType_LowPriority",
}

func (e EDRMBlobDownloadType) String() string {
	if s, ok := EDRMBlobDownloadType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EDRMBlobDownloadType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EDRMBlobDownloadErrorDetail int32

const (
	EDRMBlobDownloadErrorDetail_None                      EDRMBlobDownloadErrorDetail = 0
	EDRMBlobDownloadErrorDetail_DownloadFailed            EDRMBlobDownloadErrorDetail = 1
	EDRMBlobDownloadErrorDetail_TargetLocked              EDRMBlobDownloadErrorDetail = 2
	EDRMBlobDownloadErrorDetail_OpenZip                   EDRMBlobDownloadErrorDetail = 3
	EDRMBlobDownloadErrorDetail_ReadZipDirectory          EDRMBlobDownloadErrorDetail = 4
	EDRMBlobDownloadErrorDetail_UnexpectedZipEntry        EDRMBlobDownloadErrorDetail = 5
	EDRMBlobDownloadErrorDetail_UnzipFullFile             EDRMBlobDownloadErrorDetail = 6
	EDRMBlobDownloadErrorDetail_UnknownBlobType           EDRMBlobDownloadErrorDetail = 7
	EDRMBlobDownloadErrorDetail_UnzipStrips               EDRMBlobDownloadErrorDetail = 8
	EDRMBlobDownloadErrorDetail_UnzipMergeGuid            EDRMBlobDownloadErrorDetail = 9
	EDRMBlobDownloadErrorDetail_UnzipSignature            EDRMBlobDownloadErrorDetail = 10
	EDRMBlobDownloadErrorDetail_ApplyStrips               EDRMBlobDownloadErrorDetail = 11
	EDRMBlobDownloadErrorDetail_ApplyMergeGuid            EDRMBlobDownloadErrorDetail = 12
	EDRMBlobDownloadErrorDetail_ApplySignature            EDRMBlobDownloadErrorDetail = 13
	EDRMBlobDownloadErrorDetail_AppIdMismatch             EDRMBlobDownloadErrorDetail = 14
	EDRMBlobDownloadErrorDetail_AppIdUnexpected           EDRMBlobDownloadErrorDetail = 15
	EDRMBlobDownloadErrorDetail_AppliedSignatureCorrupt   EDRMBlobDownloadErrorDetail = 16
	EDRMBlobDownloadErrorDetail_ApplyValveSignatureHeader EDRMBlobDownloadErrorDetail = 17
	EDRMBlobDownloadErrorDetail_UnzipValveSignatureHeader EDRMBlobDownloadErrorDetail = 18
	EDRMBlobDownloadErrorDetail_PathManipulationError     EDRMBlobDownloadErrorDetail = 19
	EDRMBlobDownloadErrorDetail_TargetLocked_Base         EDRMBlobDownloadErrorDetail = 65536
	EDRMBlobDownloadErrorDetail_TargetLocked_Max          EDRMBlobDownloadErrorDetail = 131071
	EDRMBlobDownloadErrorDetail_NextBase                  EDRMBlobDownloadErrorDetail = 131072
)

var EDRMBlobDownloadErrorDetail_name = map[EDRMBlobDownloadErrorDetail]string{
	0:      "EDRMBlobDownloadErrorDetail_None",
	1:      "EDRMBlobDownloadErrorDetail_DownloadFailed",
	2:      "EDRMBlobDownloadErrorDetail_TargetLocked",
	3:      "EDRMBlobDownloadErrorDetail_OpenZip",
	4:      "EDRMBlobDownloadErrorDetail_ReadZipDirectory",
	5:      "EDRMBlobDownloadErrorDetail_UnexpectedZipEntry",
	6:      "EDRMBlobDownloadErrorDetail_UnzipFullFile",
	7:      "EDRMBlobDownloadErrorDetail_UnknownBlobType",
	8:      "EDRMBlobDownloadErrorDetail_UnzipStrips",
	9:      "EDRMBlobDownloadErrorDetail_UnzipMergeGuid",
	10:     "EDRMBlobDownloadErrorDetail_UnzipSignature",
	11:     "EDRMBlobDownloadErrorDetail_ApplyStrips",
	12:     "EDRMBlobDownloadErrorDetail_ApplyMergeGuid",
	13:     "EDRMBlobDownloadErrorDetail_ApplySignature",
	14:     "EDRMBlobDownloadErrorDetail_AppIdMismatch",
	15:     "EDRMBlobDownloadErrorDetail_AppIdUnexpected",
	16:     "EDRMBlobDownloadErrorDetail_AppliedSignatureCorrupt",
	17:     "EDRMBlobDownloadErrorDetail_ApplyValveSignatureHeader",
	18:     "EDRMBlobDownloadErrorDetail_UnzipValveSignatureHeader",
	19:     "EDRMBlobDownloadErrorDetail_PathManipulationError",
	65536:  "EDRMBlobDownloadErrorDetail_TargetLocked_Base",
	131071: "EDRMBlobDownloadErrorDetail_TargetLocked_Max",
	131072: "EDRMBlobDownloadErrorDetail_NextBase",
}

func (e EDRMBlobDownloadErrorDetail) String() string {
	if s, ok := EDRMBlobDownloadErrorDetail_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EDRMBlobDownloadErrorDetail_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EClientStat int32

const (
	EClientStat_P2PConnectionsUDP   EClientStat = 0
	EClientStat_P2PConnectionsRelay EClientStat = 1
	EClientStat_P2PGameConnections  EClientStat = 2
	EClientStat_P2PVoiceConnections EClientStat = 3
	EClientStat_BytesDownloaded     EClientStat = 4
)

var EClientStat_name = map[EClientStat]string{
	0: "EClientStat_P2PConnectionsUDP",
	1: "EClientStat_P2PConnectionsRelay",
	2: "EClientStat_P2PGameConnections",
	3: "EClientStat_P2PVoiceConnections",
	4: "EClientStat_BytesDownloaded",
}

func (e EClientStat) String() string {
	if s, ok := EClientStat_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EClientStat_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EClientStatAggregateMethod int32

const (
	EClientStatAggregateMethod_LatestOnly EClientStatAggregateMethod = 0
	EClientStatAggregateMethod_Sum        EClientStatAggregateMethod = 1
	EClientStatAggregateMethod_Event      EClientStatAggregateMethod = 2
	EClientStatAggregateMethod_Scalar     EClientStatAggregateMethod = 3
)

var EClientStatAggregateMethod_name = map[EClientStatAggregateMethod]string{
	0: "EClientStatAggregateMethod_LatestOnly",
	1: "EClientStatAggregateMethod_Sum",
	2: "EClientStatAggregateMethod_Event",
	3: "EClientStatAggregateMethod_Scalar",
}

func (e EClientStatAggregateMethod) String() string {
	if s, ok := EClientStatAggregateMethod_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EClientStatAggregateMethod_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ELeaderboardDataRequest int32

const (
	ELeaderboardDataRequest_Global           ELeaderboardDataRequest = 0
	ELeaderboardDataRequest_GlobalAroundUser ELeaderboardDataRequest = 1
	ELeaderboardDataRequest_Friends          ELeaderboardDataRequest = 2
	ELeaderboardDataRequest_Users            ELeaderboardDataRequest = 3
)

var ELeaderboardDataRequest_name = map[ELeaderboardDataRequest]string{
	0: "ELeaderboardDataRequest_Global",
	1: "ELeaderboardDataRequest_GlobalAroundUser",
	2: "ELeaderboardDataRequest_Friends",
	3: "ELeaderboardDataRequest_Users",
}

func (e ELeaderboardDataRequest) String() string {
	if s, ok := ELeaderboardDataRequest_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ELeaderboardDataRequest_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ELeaderboardSortMethod int32

const (
	ELeaderboardSortMethod_None       ELeaderboardSortMethod = 0
	ELeaderboardSortMethod_Ascending  ELeaderboardSortMethod = 1
	ELeaderboardSortMethod_Descending ELeaderboardSortMethod = 2
)

var ELeaderboardSortMethod_name = map[ELeaderboardSortMethod]string{
	0: "ELeaderboardSortMethod_None",
	1: "ELeaderboardSortMethod_Ascending",
	2: "ELeaderboardSortMethod_Descending",
}

func (e ELeaderboardSortMethod) String() string {
	if s, ok := ELeaderboardSortMethod_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ELeaderboardSortMethod_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ELeaderboardDisplayType int32

const (
	ELeaderboardDisplayType_None             ELeaderboardDisplayType = 0
	ELeaderboardDisplayType_Numeric          ELeaderboardDisplayType = 1
	ELeaderboardDisplayType_TimeSeconds      ELeaderboardDisplayType = 2
	ELeaderboardDisplayType_TimeMilliSeconds ELeaderboardDisplayType = 3
)

var ELeaderboardDisplayType_name = map[ELeaderboardDisplayType]string{
	0: "ELeaderboardDisplayType_None",
	1: "ELeaderboardDisplayType_Numeric",
	2: "ELeaderboardDisplayType_TimeSeconds",
	3: "ELeaderboardDisplayType_TimeMilliSeconds",
}

func (e ELeaderboardDisplayType) String() string {
	if s, ok := ELeaderboardDisplayType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ELeaderboardDisplayType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ELeaderboardUploadScoreMethod int32

const (
	ELeaderboardUploadScoreMethod_None        ELeaderboardUploadScoreMethod = 0
	ELeaderboardUploadScoreMethod_KeepBest    ELeaderboardUploadScoreMethod = 1
	ELeaderboardUploadScoreMethod_ForceUpdate ELeaderboardUploadScoreMethod = 2
)

var ELeaderboardUploadScoreMethod_name = map[ELeaderboardUploadScoreMethod]string{
	0: "ELeaderboardUploadScoreMethod_None",
	1: "ELeaderboardUploadScoreMethod_KeepBest",
	2: "ELeaderboardUploadScoreMethod_ForceUpdate",
}

func (e ELeaderboardUploadScoreMethod) String() string {
	if s, ok := ELeaderboardUploadScoreMethod_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ELeaderboardUploadScoreMethod_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EUCMFilePrivacyState int32

const (
	EUCMFilePrivacyState_Invalid     EUCMFilePrivacyState = -1
	EUCMFilePrivacyState_Private     EUCMFilePrivacyState = 2
	EUCMFilePrivacyState_FriendsOnly EUCMFilePrivacyState = 4
	EUCMFilePrivacyState_Public      EUCMFilePrivacyState = 8
	EUCMFilePrivacyState_All         EUCMFilePrivacyState = EUCMFilePrivacyState_Public | EUCMFilePrivacyState_FriendsOnly | EUCMFilePrivacyState_Private
)

var EUCMFilePrivacyState_name = map[EUCMFilePrivacyState]string{
	-1: "EUCMFilePrivacyState_Invalid",
	2:  "EUCMFilePrivacyState_Private",
	4:  "EUCMFilePrivacyState_FriendsOnly",
	8:  "EUCMFilePrivacyState_Public",
	14: "EUCMFilePrivacyState_All",
}

func (e EUCMFilePrivacyState) String() string {
	if s, ok := EUCMFilePrivacyState_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EUCMFilePrivacyState_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EPublishedFileQueryType int32

const (
	EPublishedFileQueryType_RankedByVote                                  EPublishedFileQueryType = 0
	EPublishedFileQueryType_RankedByPublicationDate                       EPublishedFileQueryType = 1
	EPublishedFileQueryType_AcceptedForGameRankedByAcceptanceDate         EPublishedFileQueryType = 2
	EPublishedFileQueryType_RankedByTrend                                 EPublishedFileQueryType = 3
	EPublishedFileQueryType_FavoritedByFriendsRankedByPublicationDate     EPublishedFileQueryType = 4
	EPublishedFileQueryType_CreatedByFriendsRankedByPublicationDate       EPublishedFileQueryType = 5
	EPublishedFileQueryType_RankedByNumTimesReported                      EPublishedFileQueryType = 6
	EPublishedFileQueryType_CreatedByFollowedUsersRankedByPublicationDate EPublishedFileQueryType = 7
	EPublishedFileQueryType_NotYetRated                                   EPublishedFileQueryType = 8
	EPublishedFileQueryType_RankedByTotalUniqueSubscriptions              EPublishedFileQueryType = 9
	EPublishedFileQueryType_RankedByTotalVotesAsc                         EPublishedFileQueryType = 10
	EPublishedFileQueryType_RankedByVotesUp                               EPublishedFileQueryType = 11
	EPublishedFileQueryType_RankedByTextSearch                            EPublishedFileQueryType = 12
	EPublishedFileQueryType_RankedByPlaytimeTrend                         EPublishedFileQueryType = 13
	EPublishedFileQueryType_RankedByTotalPlaytime                         EPublishedFileQueryType = 14
	EPublishedFileQueryType_RankedByAveragePlaytimeTrend                  EPublishedFileQueryType = 15
	EPublishedFileQueryType_RankedByLifetimeAveragePlaytime               EPublishedFileQueryType = 16
	EPublishedFileQueryType_RankedByPlaytimeSessionsTrend                 EPublishedFileQueryType = 17
	EPublishedFileQueryType_RankedByLifetimePlaytimeSessions              EPublishedFileQueryType = 18
	EPublishedFileQueryType_RankedByInappropriateContentRating            EPublishedFileQueryType = 19
)

var EPublishedFileQueryType_name = map[EPublishedFileQueryType]string{
	0:  "EPublishedFileQueryType_RankedByVote",
	1:  "EPublishedFileQueryType_RankedByPublicationDate",
	2:  "EPublishedFileQueryType_AcceptedForGameRankedByAcceptanceDate",
	3:  "EPublishedFileQueryType_RankedByTrend",
	4:  "EPublishedFileQueryType_FavoritedByFriendsRankedByPublicationDate",
	5:  "EPublishedFileQueryType_CreatedByFriendsRankedByPublicationDate",
	6:  "EPublishedFileQueryType_RankedByNumTimesReported",
	7:  "EPublishedFileQueryType_CreatedByFollowedUsersRankedByPublicationDate",
	8:  "EPublishedFileQueryType_NotYetRated",
	9:  "EPublishedFileQueryType_RankedByTotalUniqueSubscriptions",
	10: "EPublishedFileQueryType_RankedByTotalVotesAsc",
	11: "EPublishedFileQueryType_RankedByVotesUp",
	12: "EPublishedFileQueryType_RankedByTextSearch",
	13: "EPublishedFileQueryType_RankedByPlaytimeTrend",
	14: "EPublishedFileQueryType_RankedByTotalPlaytime",
	15: "EPublishedFileQueryType_RankedByAveragePlaytimeTrend",
	16: "EPublishedFileQueryType_RankedByLifetimeAveragePlaytime",
	17: "EPublishedFileQueryType_RankedByPlaytimeSessionsTrend",
	18: "EPublishedFileQueryType_RankedByLifetimePlaytimeSessions",
	19: "EPublishedFileQueryType_RankedByInappropriateContentRating",
}

func (e EPublishedFileQueryType) String() string {
	if s, ok := EPublishedFileQueryType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EPublishedFileQueryType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EPublishedFileInappropriateProvider int32

const (
	EPublishedFileInappropriateProvider_Invalid EPublishedFileInappropriateProvider = 0
	EPublishedFileInappropriateProvider_Google  EPublishedFileInappropriateProvider = 1
	EPublishedFileInappropriateProvider_Amazon  EPublishedFileInappropriateProvider = 2
)

var EPublishedFileInappropriateProvider_name = map[EPublishedFileInappropriateProvider]string{
	0: "EPublishedFileInappropriateProvider_Invalid",
	1: "EPublishedFileInappropriateProvider_Google",
	2: "EPublishedFileInappropriateProvider_Amazon",
}

func (e EPublishedFileInappropriateProvider) String() string {
	if s, ok := EPublishedFileInappropriateProvider_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EPublishedFileInappropriateProvider_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EPublishedFileInappropriateResult int32

const (
	EPublishedFileInappropriateResult_NotScanned   EPublishedFileInappropriateResult = 0
	EPublishedFileInappropriateResult_VeryUnlikely EPublishedFileInappropriateResult = 1
	EPublishedFileInappropriateResult_Unlikely     EPublishedFileInappropriateResult = 30
	EPublishedFileInappropriateResult_Possible     EPublishedFileInappropriateResult = 50
	EPublishedFileInappropriateResult_Likely       EPublishedFileInappropriateResult = 75
	EPublishedFileInappropriateResult_VeryLikely   EPublishedFileInappropriateResult = 100
)

var EPublishedFileInappropriateResult_name = map[EPublishedFileInappropriateResult]string{
	0:   "EPublishedFileInappropriateResult_NotScanned",
	1:   "EPublishedFileInappropriateResult_VeryUnlikely",
	30:  "EPublishedFileInappropriateResult_Unlikely",
	50:  "EPublishedFileInappropriateResult_Possible",
	75:  "EPublishedFileInappropriateResult_Likely",
	100: "EPublishedFileInappropriateResult_VeryLikely",
}

func (e EPublishedFileInappropriateResult) String() string {
	if s, ok := EPublishedFileInappropriateResult_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EPublishedFileInappropriateResult_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EDisplayStatus int32

const (
	EDisplayStatus_Invalid          EDisplayStatus = 0
	EDisplayStatus_Launching        EDisplayStatus = 1
	EDisplayStatus_Uninstalling     EDisplayStatus = 2
	EDisplayStatus_Installing       EDisplayStatus = 3
	EDisplayStatus_Running          EDisplayStatus = 4
	EDisplayStatus_Validating       EDisplayStatus = 5
	EDisplayStatus_Updating         EDisplayStatus = 6
	EDisplayStatus_Downloading      EDisplayStatus = 7
	EDisplayStatus_Synchronizing    EDisplayStatus = 8
	EDisplayStatus_ReadyToInstall   EDisplayStatus = 9
	EDisplayStatus_ReadyToPreload   EDisplayStatus = 10
	EDisplayStatus_ReadyToLaunch    EDisplayStatus = 11
	EDisplayStatus_RegionRestricted EDisplayStatus = 12
	EDisplayStatus_PresaleOnly      EDisplayStatus = 13
	EDisplayStatus_InvalidPlatform  EDisplayStatus = 14
	EDisplayStatus_ParentalBlocked  EDisplayStatus = 15
	EDisplayStatus_PreloadOnly      EDisplayStatus = 16
	EDisplayStatus_BorrowerLocked   EDisplayStatus = 17
	EDisplayStatus_UpdatePaused     EDisplayStatus = 18
	EDisplayStatus_UpdateQueued     EDisplayStatus = 19
	EDisplayStatus_UpdateRequired   EDisplayStatus = 20
	EDisplayStatus_UpdateDisabled   EDisplayStatus = 21
	EDisplayStatus_DownloadPaused   EDisplayStatus = 22
	EDisplayStatus_DownloadQueued   EDisplayStatus = 23
	EDisplayStatus_DownloadRequired EDisplayStatus = 24
	EDisplayStatus_DownloadDisabled EDisplayStatus = 25
	EDisplayStatus_LicensePending   EDisplayStatus = 26
	EDisplayStatus_LicenseExpired   EDisplayStatus = 27
	EDisplayStatus_AvailForFree     EDisplayStatus = 28
	EDisplayStatus_AvailToBorrow    EDisplayStatus = 29
	EDisplayStatus_AvailGuestPass   EDisplayStatus = 30
	EDisplayStatus_Purchase         EDisplayStatus = 31
)

var EDisplayStatus_name = map[EDisplayStatus]string{
	0:  "EDisplayStatus_Invalid",
	1:  "EDisplayStatus_Launching",
	2:  "EDisplayStatus_Uninstalling",
	3:  "EDisplayStatus_Installing",
	4:  "EDisplayStatus_Running",
	5:  "EDisplayStatus_Validating",
	6:  "EDisplayStatus_Updating",
	7:  "EDisplayStatus_Downloading",
	8:  "EDisplayStatus_Synchronizing",
	9:  "EDisplayStatus_ReadyToInstall",
	10: "EDisplayStatus_ReadyToPreload",
	11: "EDisplayStatus_ReadyToLaunch",
	12: "EDisplayStatus_RegionRestricted",
	13: "EDisplayStatus_PresaleOnly",
	14: "EDisplayStatus_InvalidPlatform",
	15: "EDisplayStatus_ParentalBlocked",
	16: "EDisplayStatus_PreloadOnly",
	17: "EDisplayStatus_BorrowerLocked",
	18: "EDisplayStatus_UpdatePaused",
	19: "EDisplayStatus_UpdateQueued",
	20: "EDisplayStatus_UpdateRequired",
	21: "EDisplayStatus_UpdateDisabled",
	22: "EDisplayStatus_DownloadPaused",
	23: "EDisplayStatus_DownloadQueued",
	24: "EDisplayStatus_DownloadRequired",
	25: "EDisplayStatus_DownloadDisabled",
	26: "EDisplayStatus_LicensePending",
	27: "EDisplayStatus_LicenseExpired",
	28: "EDisplayStatus_AvailForFree",
	29: "EDisplayStatus_AvailToBorrow",
	30: "EDisplayStatus_AvailGuestPass",
	31: "EDisplayStatus_Purchase",
}

func (e EDisplayStatus) String() string {
	if s, ok := EDisplayStatus_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EDisplayStatus_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EAppType int32

const (
	EAppType_Invalid     EAppType = 0
	EAppType_Game        EAppType = 1
	EAppType_Application EAppType = 2
	EAppType_Tool        EAppType = 4
	EAppType_Demo        EAppType = 8
	EAppType_Deprected   EAppType = 16
	EAppType_DLC         EAppType = 32
	EAppType_Guide       EAppType = 64
	EAppType_Driver      EAppType = 128
	EAppType_Config      EAppType = 256
	EAppType_Hardware    EAppType = 512
	EAppType_Franchise   EAppType = 1024
	EAppType_Video       EAppType = 2048
	EAppType_Plugin      EAppType = 4096
	EAppType_Music       EAppType = 8192
	EAppType_Series      EAppType = 16384
	EAppType_Comic       EAppType = 32768
	EAppType_Beta        EAppType = 65536
	EAppType_Shortcut    EAppType = 1073741824
	EAppType_DepotOnly   EAppType = -2147483648
)

var EAppType_name = map[EAppType]string{
	0:           "EAppType_Invalid",
	1:           "EAppType_Game",
	2:           "EAppType_Application",
	4:           "EAppType_Tool",
	8:           "EAppType_Demo",
	16:          "EAppType_Deprected",
	32:          "EAppType_DLC",
	64:          "EAppType_Guide",
	128:         "EAppType_Driver",
	256:         "EAppType_Config",
	512:         "EAppType_Hardware",
	1024:        "EAppType_Franchise",
	2048:        "EAppType_Video",
	4096:        "EAppType_Plugin",
	8192:        "EAppType_Music",
	16384:       "EAppType_Series",
	32768:       "EAppType_Comic",
	65536:       "EAppType_Beta",
	1073741824:  "EAppType_Shortcut",
	-2147483648: "EAppType_DepotOnly",
}

func (e EAppType) String() string {
	if s, ok := EAppType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EAppType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatRoomGroupType int32

const (
	EChatRoomGroupType_Default     EChatRoomGroupType = 0
	EChatRoomGroupType_Unmoderated EChatRoomGroupType = 1
)

var EChatRoomGroupType_name = map[EChatRoomGroupType]string{
	0: "EChatRoomGroupType_Default",
	1: "EChatRoomGroupType_Unmoderated",
}

func (e EChatRoomGroupType) String() string {
	if s, ok := EChatRoomGroupType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatRoomGroupType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatroomNotificationLevel int32

const (
	EChatroomNotificationLevel_Invalid     EChatroomNotificationLevel = 0
	EChatroomNotificationLevel_None        EChatroomNotificationLevel = 1
	EChatroomNotificationLevel_MentionMe   EChatroomNotificationLevel = 2
	EChatroomNotificationLevel_MentionAll  EChatroomNotificationLevel = 3
	EChatroomNotificationLevel_AllMessages EChatroomNotificationLevel = 4
)

var EChatroomNotificationLevel_name = map[EChatroomNotificationLevel]string{
	0: "EChatroomNotificationLevel_Invalid",
	1: "EChatroomNotificationLevel_None",
	2: "EChatroomNotificationLevel_MentionMe",
	3: "EChatroomNotificationLevel_MentionAll",
	4: "EChatroomNotificationLevel_AllMessages",
}

func (e EChatroomNotificationLevel) String() string {
	if s, ok := EChatroomNotificationLevel_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatroomNotificationLevel_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatRoomMemberStateChange int32

const (
	EChatRoomMemberStateChange_Invalid         EChatRoomMemberStateChange = 0
	EChatRoomMemberStateChange_Joined          EChatRoomMemberStateChange = 1
	EChatRoomMemberStateChange_Parted          EChatRoomMemberStateChange = 2
	EChatRoomMemberStateChange_Kicked          EChatRoomMemberStateChange = 3
	EChatRoomMemberStateChange_Invited         EChatRoomMemberStateChange = 4
	EChatRoomMemberStateChange_RankChanged     EChatRoomMemberStateChange = 7
	EChatRoomMemberStateChange_InviteDismissed EChatRoomMemberStateChange = 8
	EChatRoomMemberStateChange_Muted           EChatRoomMemberStateChange = 9
	EChatRoomMemberStateChange_Banned          EChatRoomMemberStateChange = 10
	EChatRoomMemberStateChange_RolesChanged    EChatRoomMemberStateChange = 12
)

var EChatRoomMemberStateChange_name = map[EChatRoomMemberStateChange]string{
	0:  "EChatRoomMemberStateChange_Invalid",
	1:  "EChatRoomMemberStateChange_Joined",
	2:  "EChatRoomMemberStateChange_Parted",
	3:  "EChatRoomMemberStateChange_Kicked",
	4:  "EChatRoomMemberStateChange_Invited",
	7:  "EChatRoomMemberStateChange_RankChanged",
	8:  "EChatRoomMemberStateChange_InviteDismissed",
	9:  "EChatRoomMemberStateChange_Muted",
	10: "EChatRoomMemberStateChange_Banned",
	12: "EChatRoomMemberStateChange_RolesChanged",
}

func (e EChatRoomMemberStateChange) String() string {
	if s, ok := EChatRoomMemberStateChange_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatRoomMemberStateChange_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatRoomServerMsg int32

const (
	EChatRoomServerMsg_Invalid                EChatRoomServerMsg = 0
	EChatRoomServerMsg_RenameChatRoom         EChatRoomServerMsg = 1
	EChatRoomServerMsg_Joined                 EChatRoomServerMsg = 2
	EChatRoomServerMsg_Parted                 EChatRoomServerMsg = 3
	EChatRoomServerMsg_Kicked                 EChatRoomServerMsg = 4
	EChatRoomServerMsg_Invited                EChatRoomServerMsg = 5
	EChatRoomServerMsg_InviteDismissed        EChatRoomServerMsg = 8
	EChatRoomServerMsg_ChatRoomTaglineChanged EChatRoomServerMsg = 9
	EChatRoomServerMsg_ChatRoomAvatarChanged  EChatRoomServerMsg = 10
	EChatRoomServerMsg_AppCustom              EChatRoomServerMsg = 11
)

var EChatRoomServerMsg_name = map[EChatRoomServerMsg]string{
	0:  "EChatRoomServerMsg_Invalid",
	1:  "EChatRoomServerMsg_RenameChatRoom",
	2:  "EChatRoomServerMsg_Joined",
	3:  "EChatRoomServerMsg_Parted",
	4:  "EChatRoomServerMsg_Kicked",
	5:  "EChatRoomServerMsg_Invited",
	8:  "EChatRoomServerMsg_InviteDismissed",
	9:  "EChatRoomServerMsg_ChatRoomTaglineChanged",
	10: "EChatRoomServerMsg_ChatRoomAvatarChanged",
	11: "EChatRoomServerMsg_AppCustom",
}

func (e EChatRoomServerMsg) String() string {
	if s, ok := EChatRoomServerMsg_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatRoomServerMsg_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatRoomGroupRank int32

const (
	EChatRoomGroupRank_Default   EChatRoomGroupRank = 0
	EChatRoomGroupRank_Viewer    EChatRoomGroupRank = 10
	EChatRoomGroupRank_Guest     EChatRoomGroupRank = 15
	EChatRoomGroupRank_Member    EChatRoomGroupRank = 20
	EChatRoomGroupRank_Moderator EChatRoomGroupRank = 30
	EChatRoomGroupRank_Officer   EChatRoomGroupRank = 40
	EChatRoomGroupRank_Owner     EChatRoomGroupRank = 50
)

var EChatRoomGroupRank_name = map[EChatRoomGroupRank]string{
	0:  "EChatRoomGroupRank_Default",
	10: "EChatRoomGroupRank_Viewer",
	15: "EChatRoomGroupRank_Guest",
	20: "EChatRoomGroupRank_Member",
	30: "EChatRoomGroupRank_Moderator",
	40: "EChatRoomGroupRank_Officer",
	50: "EChatRoomGroupRank_Owner",
}

func (e EChatRoomGroupRank) String() string {
	if s, ok := EChatRoomGroupRank_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatRoomGroupRank_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatRoomGroupPermissions int32

const (
	EChatRoomGroupPermissions_Default         EChatRoomGroupPermissions = 0
	EChatRoomGroupPermissions_Valid           EChatRoomGroupPermissions = 1
	EChatRoomGroupPermissions_CanInvite       EChatRoomGroupPermissions = 2
	EChatRoomGroupPermissions_CanKick         EChatRoomGroupPermissions = 4
	EChatRoomGroupPermissions_CanBan          EChatRoomGroupPermissions = 8
	EChatRoomGroupPermissions_CanAdminChannel EChatRoomGroupPermissions = 16
)

var EChatRoomGroupPermissions_name = map[EChatRoomGroupPermissions]string{
	0:  "EChatRoomGroupPermissions_Default",
	1:  "EChatRoomGroupPermissions_Valid",
	2:  "EChatRoomGroupPermissions_CanInvite",
	4:  "EChatRoomGroupPermissions_CanKick",
	8:  "EChatRoomGroupPermissions_CanBan",
	16: "EChatRoomGroupPermissions_CanAdminChannel",
}

func (e EChatRoomGroupPermissions) String() string {
	if s, ok := EChatRoomGroupPermissions_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatRoomGroupPermissions_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatRoomGroupAction int32

const (
	EChatRoomGroupAction_Default                   EChatRoomGroupAction = 0
	EChatRoomGroupAction_CreateRenameDeleteChannel EChatRoomGroupAction = 1
	EChatRoomGroupAction_Kick                      EChatRoomGroupAction = 2
	EChatRoomGroupAction_Ban                       EChatRoomGroupAction = 3
	EChatRoomGroupAction_Invite                    EChatRoomGroupAction = 4
	EChatRoomGroupAction_ChangeTaglineAvatarName   EChatRoomGroupAction = 5
	EChatRoomGroupAction_Chat                      EChatRoomGroupAction = 6
	EChatRoomGroupAction_ViewHistory               EChatRoomGroupAction = 7
	EChatRoomGroupAction_ChangeGroupRoles          EChatRoomGroupAction = 8
	EChatRoomGroupAction_ChangeUserRoles           EChatRoomGroupAction = 9
	EChatRoomGroupAction_MentionAll                EChatRoomGroupAction = 10
	EChatRoomGroupAction_SetWatchingBroadcast      EChatRoomGroupAction = 11
)

var EChatRoomGroupAction_name = map[EChatRoomGroupAction]string{
	0:  "EChatRoomGroupAction_Default",
	1:  "EChatRoomGroupAction_CreateRenameDeleteChannel",
	2:  "EChatRoomGroupAction_Kick",
	3:  "EChatRoomGroupAction_Ban",
	4:  "EChatRoomGroupAction_Invite",
	5:  "EChatRoomGroupAction_ChangeTaglineAvatarName",
	6:  "EChatRoomGroupAction_Chat",
	7:  "EChatRoomGroupAction_ViewHistory",
	8:  "EChatRoomGroupAction_ChangeGroupRoles",
	9:  "EChatRoomGroupAction_ChangeUserRoles",
	10: "EChatRoomGroupAction_MentionAll",
	11: "EChatRoomGroupAction_SetWatchingBroadcast",
}

func (e EChatRoomGroupAction) String() string {
	if s, ok := EChatRoomGroupAction_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatRoomGroupAction_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EChatRoomJoinState int32

const (
	EChatRoomJoinState_Default EChatRoomJoinState = 0
	EChatRoomJoinState_None    EChatRoomJoinState = 1
	EChatRoomJoinState_Joined  EChatRoomJoinState = 2
)

var EChatRoomJoinState_name = map[EChatRoomJoinState]string{
	0: "EChatRoomJoinState_Default",
	1: "EChatRoomJoinState_None",
	2: "EChatRoomJoinState_Joined",
}

func (e EChatRoomJoinState) String() string {
	if s, ok := EChatRoomJoinState_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EChatRoomJoinState_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EVoiceCallState int32

const (
	EVoiceCallState_None                                 EVoiceCallState = 0
	EVoiceCallState_ScheduledInitiate                    EVoiceCallState = 1
	EVoiceCallState_RequestedMicAccess                   EVoiceCallState = 2
	EVoiceCallState_LocalMicOnly                         EVoiceCallState = 3
	EVoiceCallState_CreatePeerConnection                 EVoiceCallState = 4
	EVoiceCallState_InitatedWebRTCSession                EVoiceCallState = 5
	EVoiceCallState_WebRTCConnectedWaitingOnIceConnected EVoiceCallState = 6
	EVoiceCallState_RequestedPermission                  EVoiceCallState = 7
	EVoiceCallState_NotifyingVoiceChatOfWebRTCSession    EVoiceCallState = 8
	EVoiceCallState_Connected                            EVoiceCallState = 9
)

var EVoiceCallState_name = map[EVoiceCallState]string{
	0: "EVoiceCallState_None",
	1: "EVoiceCallState_ScheduledInitiate",
	2: "EVoiceCallState_RequestedMicAccess",
	3: "EVoiceCallState_LocalMicOnly",
	4: "EVoiceCallState_CreatePeerConnection",
	5: "EVoiceCallState_InitatedWebRTCSession",
	6: "EVoiceCallState_WebRTCConnectedWaitingOnIceConnected",
	7: "EVoiceCallState_RequestedPermission",
	8: "EVoiceCallState_NotifyingVoiceChatOfWebRTCSession",
	9: "EVoiceCallState_Connected",
}

func (e EVoiceCallState) String() string {
	if s, ok := EVoiceCallState_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EVoiceCallState_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ETradeOfferState int32

const (
	ETradeOfferState_Invalid                  ETradeOfferState = 1
	ETradeOfferState_Active                   ETradeOfferState = 2
	ETradeOfferState_Accepted                 ETradeOfferState = 3
	ETradeOfferState_Countered                ETradeOfferState = 4
	ETradeOfferState_Expired                  ETradeOfferState = 5
	ETradeOfferState_Canceled                 ETradeOfferState = 6
	ETradeOfferState_Declined                 ETradeOfferState = 7
	ETradeOfferState_InvalidItems             ETradeOfferState = 8
	ETradeOfferState_CreatedNeedsConfirmation ETradeOfferState = 9
	ETradeOfferState_CanceledBySecondFactor   ETradeOfferState = 10
	ETradeOfferState_InEscrow                 ETradeOfferState = 11
)

var ETradeOfferState_name = map[ETradeOfferState]string{
	1:  "ETradeOfferState_Invalid",
	2:  "ETradeOfferState_Active",
	3:  "ETradeOfferState_Accepted",
	4:  "ETradeOfferState_Countered",
	5:  "ETradeOfferState_Expired",
	6:  "ETradeOfferState_Canceled",
	7:  "ETradeOfferState_Declined",
	8:  "ETradeOfferState_InvalidItems",
	9:  "ETradeOfferState_CreatedNeedsConfirmation",
	10: "ETradeOfferState_CanceledBySecondFactor",
	11: "ETradeOfferState_InEscrow",
}

func (e ETradeOfferState) String() string {
	if s, ok := ETradeOfferState_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ETradeOfferState_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ETradeOfferConfirmationMethod int32

const (
	ETradeOfferConfirmationMethod_Invalid   ETradeOfferConfirmationMethod = 0
	ETradeOfferConfirmationMethod_Email     ETradeOfferConfirmationMethod = 1
	ETradeOfferConfirmationMethod_MobileApp ETradeOfferConfirmationMethod = 2
)

var ETradeOfferConfirmationMethod_name = map[ETradeOfferConfirmationMethod]string{
	0: "ETradeOfferConfirmationMethod_Invalid",
	1: "ETradeOfferConfirmationMethod_Email",
	2: "ETradeOfferConfirmationMethod_MobileApp",
}

func (e ETradeOfferConfirmationMethod) String() string {
	if s, ok := ETradeOfferConfirmationMethod_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ETradeOfferConfirmationMethod_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ELobbyType int32

const (
	ELobbyType_Private       ELobbyType = 0
	ELobbyType_FriendsOnly   ELobbyType = 1
	ELobbyType_Public        ELobbyType = 2
	ELobbyType_Invisible     ELobbyType = 3
	ELobbyType_PrivateUnique ELobbyType = 4
)

var ELobbyType_name = map[ELobbyType]string{
	0: "ELobbyType_Private",
	1: "ELobbyType_FriendsOnly",
	2: "ELobbyType_Public",
	3: "ELobbyType_Invisible",
	4: "ELobbyType_PrivateUnique",
}

func (e ELobbyType) String() string {
	if s, ok := ELobbyType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ELobbyType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ELobbyFilterType int32

const (
	ELobbyFilterType_String         ELobbyFilterType = 0
	ELobbyFilterType_Numerical      ELobbyFilterType = 1
	ELobbyFilterType_SlotsAvailable ELobbyFilterType = 2
	ELobbyFilterType_NearValue      ELobbyFilterType = 3
	ELobbyFilterType_Distance       ELobbyFilterType = 4
)

var ELobbyFilterType_name = map[ELobbyFilterType]string{
	0: "ELobbyFilterType_String",
	1: "ELobbyFilterType_Numerical",
	2: "ELobbyFilterType_SlotsAvailable",
	3: "ELobbyFilterType_NearValue",
	4: "ELobbyFilterType_Distance",
}

func (e ELobbyFilterType) String() string {
	if s, ok := ELobbyFilterType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ELobbyFilterType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ELobbyComparison int32

const (
	ELobbyComparison_EqualToOrLessThan    ELobbyComparison = -2
	ELobbyComparison_LessThan             ELobbyComparison = -1
	ELobbyComparison_Equal                ELobbyComparison = 0
	ELobbyComparison_GreaterThan          ELobbyComparison = 1
	ELobbyComparison_EqualToOrGreaterThan ELobbyComparison = 2
	ELobbyComparison_NotEqual             ELobbyComparison = 3
)

var ELobbyComparison_name = map[ELobbyComparison]string{
	-2: "ELobbyComparison_EqualToOrLessThan",
	-1: "ELobbyComparison_LessThan",
	0:  "ELobbyComparison_Equal",
	1:  "ELobbyComparison_GreaterThan",
	2:  "ELobbyComparison_EqualToOrGreaterThan",
	3:  "ELobbyComparison_NotEqual",
}

func (e ELobbyComparison) String() string {
	if s, ok := ELobbyComparison_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ELobbyComparison_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ELobbyDistanceFilter int32

const (
	ELobbyDistanceFilter_Close     ELobbyDistanceFilter = 0
	ELobbyDistanceFilter_Default   ELobbyDistanceFilter = 1
	ELobbyDistanceFilter_Far       ELobbyDistanceFilter = 2
	ELobbyDistanceFilter_Worldwide ELobbyDistanceFilter = 3
)

var ELobbyDistanceFilter_name = map[ELobbyDistanceFilter]string{
	0: "ELobbyDistanceFilter_Close",
	1: "ELobbyDistanceFilter_Default",
	2: "ELobbyDistanceFilter_Far",
	3: "ELobbyDistanceFilter_Worldwide",
}

func (e ELobbyDistanceFilter) String() string {
	if s, ok := ELobbyDistanceFilter_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ELobbyDistanceFilter_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ESteamIPv6ConnectivityProtocol int32

const (
	ESteamIPv6ConnectivityProtocol_Invalid ESteamIPv6ConnectivityProtocol = 0
	ESteamIPv6ConnectivityProtocol_Http    ESteamIPv6ConnectivityProtocol = 1
	ESteamIPv6ConnectivityProtocol_Udp     ESteamIPv6ConnectivityProtocol = 2
)

var ESteamIPv6ConnectivityProtocol_name = map[ESteamIPv6ConnectivityProtocol]string{
	0: "ESteamIPv6ConnectivityProtocol_Invalid",
	1: "ESteamIPv6ConnectivityProtocol_Http",
	2: "ESteamIPv6ConnectivityProtocol_Udp",
}

func (e ESteamIPv6ConnectivityProtocol) String() string {
	if s, ok := ESteamIPv6ConnectivityProtocol_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ESteamIPv6ConnectivityProtocol_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ESteamIPv6ConnectivityState int32

const (
	ESteamIPv6ConnectivityState_Unknown ESteamIPv6ConnectivityState = 0
	ESteamIPv6ConnectivityState_Good    ESteamIPv6ConnectivityState = 1
	ESteamIPv6ConnectivityState_Bad     ESteamIPv6ConnectivityState = 2
)

var ESteamIPv6ConnectivityState_name = map[ESteamIPv6ConnectivityState]string{
	0: "ESteamIPv6ConnectivityState_Unknown",
	1: "ESteamIPv6ConnectivityState_Good",
	2: "ESteamIPv6ConnectivityState_Bad",
}

func (e ESteamIPv6ConnectivityState) String() string {
	if s, ok := ESteamIPv6ConnectivityState_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ESteamIPv6ConnectivityState_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ESteamRealm int32

const (
	ESteamRealm_Unknown     ESteamRealm = 0
	ESteamRealm_SteamGlobal ESteamRealm = 1
	ESteamRealm_SteamChina  ESteamRealm = 2
)

var ESteamRealm_name = map[ESteamRealm]string{
	0: "ESteamRealm_Unknown",
	1: "ESteamRealm_SteamGlobal",
	2: "ESteamRealm_SteamChina",
}

func (e ESteamRealm) String() string {
	if s, ok := ESteamRealm_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ESteamRealm_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type ELauncherType int32

const (
	ELauncherType_Default      ELauncherType = 0
	ELauncherType_PerfectWorld ELauncherType = 1
	ELauncherType_Nexon        ELauncherType = 2
	ELauncherType_CmdLine      ELauncherType = 3
	ELauncherType_CSGO         ELauncherType = 4
	ELauncherType_ClientUI     ELauncherType = 5
	ELauncherType_Headless     ELauncherType = 6
	ELauncherType_SteamChina   ELauncherType = 7
	ELauncherType_SingleApp    ELauncherType = 8
)

var ELauncherType_name = map[ELauncherType]string{
	0: "ELauncherType_Default",
	1: "ELauncherType_PerfectWorld",
	2: "ELauncherType_Nexon",
	3: "ELauncherType_CmdLine",
	4: "ELauncherType_CSGO",
	5: "ELauncherType_ClientUI",
	6: "ELauncherType_Headless",
	7: "ELauncherType_SteamChina",
	8: "ELauncherType_SingleApp",
}

func (e ELauncherType) String() string {
	if s, ok := ELauncherType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range ELauncherType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}

type EUdpPacketType uint8

const (
	EUdpPacketType_Invalid      EUdpPacketType = 0
	EUdpPacketType_ChallengeReq EUdpPacketType = 1
	EUdpPacketType_Challenge    EUdpPacketType = 2
	EUdpPacketType_Connect      EUdpPacketType = 3
	EUdpPacketType_Accept       EUdpPacketType = 4
	EUdpPacketType_Disconnect   EUdpPacketType = 5
	EUdpPacketType_Data         EUdpPacketType = 6
	EUdpPacketType_Datagram     EUdpPacketType = 7
	EUdpPacketType_Max          EUdpPacketType = 8
)

var EUdpPacketType_name = map[EUdpPacketType]string{
	0: "EUdpPacketType_Invalid",
	1: "EUdpPacketType_ChallengeReq",
	2: "EUdpPacketType_Challenge",
	3: "EUdpPacketType_Connect",
	4: "EUdpPacketType_Accept",
	5: "EUdpPacketType_Disconnect",
	6: "EUdpPacketType_Data",
	7: "EUdpPacketType_Datagram",
	8: "EUdpPacketType_Max",
}

func (e EUdpPacketType) String() string {
	if s, ok := EUdpPacketType_name[e]; ok {
		return s
	}
	var flags []string
	for k, v := range EUdpPacketType_name {
		if e&k != 0 {
			flags = append(flags, v)
		}
	}
	if len(flags) == 0 {
		return fmt.Sprintf("%d", e)
	}
	sort.Strings(flags)
	return strings.Join(flags, " | ")
}
