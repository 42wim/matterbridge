// Package token contains maps of predefined tokens that WhatsApp's binary XML
// encoding uses to save bytes when sending commonly used strings.
package token

import "fmt"

// All the currently known string tokens.
var (
	SingleByteTokens = [...]string{"", "xmlstreamstart", "xmlstreamend", "s.whatsapp.net", "type", "participant", "from", "receipt", "id", "notification", "disappearing_mode", "status", "jid", "broadcast", "user", "devices", "device_hash", "to", "offline", "message", "result", "class", "xmlns", "duration", "notify", "iq", "t", "ack", "g.us", "enc", "urn:xmpp:whatsapp:push", "presence", "config_value", "picture", "verified_name", "config_code", "key-index-list", "contact", "mediatype", "routing_info", "edge_routing", "get", "read", "urn:xmpp:ping", "fallback_hostname", "0", "chatstate", "business_hours_config", "unavailable", "download_buckets", "skmsg", "verified_level", "composing", "handshake", "device-list", "media", "text", "fallback_ip4", "media_conn", "device", "creation", "location", "config", "item", "fallback_ip6", "count", "w:profile:picture", "image", "business", "2", "hostname", "call-creator", "display_name", "relaylatency", "platform", "abprops", "success", "msg", "offline_preview", "prop", "key-index", "v", "day_of_week", "pkmsg", "version", "1", "ping", "w:p", "download", "video", "set", "specific_hours", "props", "primary", "unknown", "hash", "commerce_experience", "last", "subscribe", "max_buckets", "call", "profile", "member_since_text", "close_time", "call-id", "sticker", "mode", "participants", "value", "query", "profile_options", "open_time", "code", "list", "host", "ts", "contacts", "upload", "lid", "preview", "update", "usync", "w:stats", "delivery", "auth_ttl", "context", "fail", "cart_enabled", "appdata", "category", "atn", "direct_connection", "decrypt-fail", "relay_id", "mmg-fallback.whatsapp.net", "target", "available", "name", "last_id", "mmg.whatsapp.net", "categories", "401", "is_new", "index", "tctoken", "ip4", "token_id", "latency", "recipient", "edit", "ip6", "add", "thumbnail-document", "26", "paused", "true", "identity", "stream:error", "key", "sidelist", "background", "audio", "3", "thumbnail-image", "biz-cover-photo", "cat", "gcm", "thumbnail-video", "error", "auth", "deny", "serial", "in", "registration", "thumbnail-link", "remove", "00", "gif", "thumbnail-gif", "tag", "capability", "multicast", "item-not-found", "description", "business_hours", "config_expo_key", "md-app-state", "expiration", "fallback", "ttl", "300", "md-msg-hist", "device_orientation", "out", "w:m", "open_24h", "side_list", "token", "inactive", "01", "document", "te2", "played", "encrypt", "msgr", "hide", "direct_path", "12", "state", "not-authorized", "url", "terminate", "signature", "status-revoke-delay", "02", "te", "linked_accounts", "trusted_contact", "timezone", "ptt", "kyc-id", "privacy_token", "readreceipts", "appointment_only", "address", "expected_ts", "privacy", "7", "android", "interactive", "device-identity", "enabled", "attribute_padding", "1080", "03", "screen_height"}
	DoubleByteTokens = [...][]string{
		{"read-self", "active", "fbns", "protocol", "reaction", "screen_width", "heartbeat", "deviceid", "2:47DEQpj8", "uploadfieldstat", "voip_settings", "retry", "priority", "longitude", "conflict", "false", "ig_professional", "replaced", "preaccept", "cover_photo", "uncompressed", "encopt", "ppic", "04", "passive", "status-revoke-drop", "keygen", "540", "offer", "rate", "opus", "latitude", "w:gp2", "ver", "4", "business_profile", "medium", "sender", "prev_v_id", "email", "website", "invited", "sign_credential", "05", "transport", "skey", "reason", "peer_abtest_bucket", "America/Sao_Paulo", "appid", "refresh", "100", "06", "404", "101", "104", "107", "102", "109", "103", "member_add_mode", "105", "transaction-id", "110", "106", "outgoing", "108", "111", "tokens", "followers", "ig_handle", "self_pid", "tue", "dec", "thu", "joinable", "peer_pid", "mon", "features", "wed", "peer_device_presence", "pn", "delete", "07", "fri", "audio_duration", "admin", "connected", "delta", "rcat", "disable", "collection", "08", "480", "sat", "phash", "all", "invite", "accept", "critical_unblock_low", "group_update", "signed_credential", "blinded_credential", "eph_setting", "net", "09", "background_location", "refresh_id", "Asia/Kolkata", "privacy_mode_ts", "account_sync", "voip_payload_type", "service_areas", "acs_public_key", "v_id", "0a", "fallback_class", "relay", "actual_actors", "metadata", "w:biz", "5", "connected-limit", "notice", "0b", "host_storage", "fb_page", "subject", "privatestats", "invis", "groupadd", "010", "note.m4r", "uuid", "0c", "8000", "sun", "372", "1020", "stage", "1200", "720", "canonical", "fb", "011", "video_duration", "0d", "1140", "superadmin", "012", "Opening.m4r", "keystore_attestation", "dleq_proof", "013", "timestamp", "ab_key", "w:sync:app:state", "0e", "vertical", "600", "p_v_id", "6", "likes", "014", "500", "1260", "creator", "0f", "rte", "destination", "group", "group_info", "syncd_anti_tampering_fatal_exception_enabled", "015", "dl_bw", "Asia/Jakarta", "vp8/h.264", "online", "1320", "fb:multiway", "10", "timeout", "016", "nse_retry", "urn:xmpp:whatsapp:dirty", "017", "a_v_id", "web_shops_chat_header_button_enabled", "nse_call", "inactive-upgrade", "none", "web", "groups", "2250", "mms_hot_content_timespan_in_seconds", "contact_blacklist", "nse_read", "suspended_group_deletion_notification", "binary_version", "018", "https://www.whatsapp.com/otp/copy/", "reg_push", "shops_hide_catalog_attachment_entrypoint", "server_sync", ".", "ephemeral_messages_allowed_values", "019", "mms_vcache_aggregation_enabled", "iphone", "America/Argentina/Buenos_Aires", "01a", "mms_vcard_autodownload_size_kb", "nse_ver", "shops_header_dropdown_menu_item", "dhash", "catalog_status", "communities_mvp_new_iqs_serverprop", "blocklist", "default", "11", "ephemeral_messages_enabled", "01b", "original_dimensions", "8", "mms4_media_retry_notification_encryption_enabled", "mms4_server_error_receipt_encryption_enabled", "original_image_url", "sync", "multiway", "420", "companion_enc_static", "shops_profile_drawer_entrypoint", "01c", "vcard_as_document_size_kb", "status_video_max_duration", "request_image_url", "01d", "regular_high", "s_t", "abt", "share_ext_min_preliminary_image_quality", "01e", "32", "syncd_key_rotation_enabled", "data_namespace", "md_downgrade_read_receipts2", "patch", "polltype", "ephemeral_messages_setting", "userrate", "15", "partial_pjpeg_bw_threshold", "played-self", "catalog_exists", "01f", "mute_v2"},
		{"reject", "dirty", "announcement", "020", "13", "9", "status_video_max_bitrate", "fb:thrift_iq", "offline_batch", "022", "full", "ctwa_first_business_reply_logging", "h.264", "smax_id", "group_description_length", "https://www.whatsapp.com/otp/code", "status_image_max_edge", "smb_upsell_business_profile_enabled", "021", "web_upgrade_to_md_modal", "14", "023", "s_o", "smaller_video_thumbs_status_enabled", "media_max_autodownload", "960", "blocking_status", "peer_msg", "joinable_group_call_client_version", "group_call_video_maximization_enabled", "return_snapshot", "high", "America/Mexico_City", "entry_point_block_logging_enabled", "pop", "024", "1050", "16", "1380", "one_tap_calling_in_group_chat_size", "regular_low", "inline_joinable_education_enabled", "hq_image_max_edge", "locked", "America/Bogota", "smb_biztools_deeplink_enabled", "status_image_quality", "1088", "025", "payments_upi_intent_transaction_limit", "voip", "w:g2", "027", "md_pin_chat_enabled", "026", "multi_scan_pjpeg_download_enabled", "shops_product_grid", "transaction_id", "ctwa_context_enabled", "20", "fna", "hq_image_quality", "alt_jpeg_doc_detection_quality", "group_call_max_participants", "pkey", "America/Belem", "image_max_kbytes", "web_cart_v1_1_order_message_changes_enabled", "ctwa_context_enterprise_enabled", "urn:xmpp:whatsapp:account", "840", "Asia/Kuala_Lumpur", "max_participants", "video_remux_after_repair_enabled", "stella_addressbook_restriction_type", "660", "900", "780", "context_menu_ios13_enabled", "mute-state", "ref", "payments_request_messages", "029", "frskmsg", "vcard_max_size_kb", "sample_buffer_gif_player_enabled", "match_last_seen", "510", "4983", "video_max_bitrate", "028", "w:comms:chat", "17", "frequently_forwarded_max", "groups_privacy_blacklist", "Asia/Karachi", "02a", "web_download_document_thumb_mms_enabled", "02b", "hist_sync", "biz_block_reasons_version", "1024", "18", "web_is_direct_connection_for_plm_transparent", "view_once_write", "file_max_size", "paid_convo_id", "online_privacy_setting", "video_max_edge", "view_once_read", "enhanced_storage_management", "multi_scan_pjpeg_encoding_enabled", "ctwa_context_forward_enabled", "video_transcode_downgrade_enable", "template_doc_mime_types", "hq_image_bw_threshold", "30", "body", "u_aud_limit_sil_restarts_ctrl", "other", "participating", "w:biz:directory", "1110", "vp8", "4018", "meta", "doc_detection_image_max_edge", "image_quality", "1170", "02c", "smb_upsell_chat_banner_enabled", "key_expiry_time_second", "pid", "stella_interop_enabled", "19", "linked_device_max_count", "md_device_sync_enabled", "02d", "02e", "360", "enhanced_block_enabled", "ephemeral_icon_in_forwarding", "paid_convo_status", "gif_provider", "project_name", "server-error", "canonical_url_validation_enabled", "wallpapers_v2", "syncd_clear_chat_delete_chat_enabled", "medianotify", "02f", "shops_required_tos_version", "vote", "reset_skey_on_id_change", "030", "image_max_edge", "multicast_limit_global", "ul_bw", "21", "25", "5000", "poll", "570", "22", "031", "1280", "WhatsApp", "032", "bloks_shops_enabled", "50", "upload_host_switching_enabled", "web_ctwa_context_compose_enabled", "ptt_forwarded_features_enabled", "unblocked", "partial_pjpeg_enabled", "fbid:devices", "height", "ephemeral_group_query_ts", "group_join_permissions", "order", "033", "alt_jpeg_status_quality", "migrate", "popular-bank", "win_uwp_deprecation_killswitch_enabled", "web_download_status_thumb_mms_enabled", "blocking", "url_text", "035", "web_forwarding_limit_to_groups", "1600", "val", "1000", "syncd_msg_date_enabled", "bank-ref-id", "max_subject", "payments_web_enabled", "web_upload_document_thumb_mms_enabled", "size", "request", "ephemeral", "24", "receipt_agg", "ptt_remember_play_position", "sampling_weight", "enc_rekey", "mute_always", "037", "034", "23", "036", "action", "click_to_chat_qr_enabled", "width", "disabled", "038", "md_blocklist_v2", "played_self_enabled", "web_buttons_message_enabled", "flow_id", "clear", "450", "fbid:thread", "bloks_session_state", "America/Lima", "attachment_picker_refresh", "download_host_switching_enabled", "1792", "u_aud_limit_sil_restarts_test2", "custom_urls", "device_fanout", "optimistic_upload", "2000", "key_cipher_suite", "web_smb_upsell_in_biz_profile_enabled", "e", "039", "siri_post_status_shortcut", "pair-device", "lg", "lc", "stream_attribution_url", "model", "mspjpeg_phash_gen", "catalog_send_all", "new_multi_vcards_ui", "share_biz_vcard_enabled", "-", "clean", "200", "md_blocklist_v2_server", "03b", "03a", "web_md_migration_experience", "ptt_conversation_waveform", "u_aud_limit_sil_restarts_test1"},
		{"64", "ptt_playback_speed_enabled", "web_product_list_message_enabled", "paid_convo_ts", "27", "manufacturer", "psp-routing", "grp_uii_cleanup", "ptt_draft_enabled", "03c", "business_initiated", "web_catalog_products_onoff", "web_upload_link_thumb_mms_enabled", "03e", "mediaretry", "35", "hfm_string_changes", "28", "America/Fortaleza", "max_keys", "md_mhfs_days", "streaming_upload_chunk_size", "5541", "040", "03d", "2675", "03f", "...", "512", "mute", "48", "041", "alt_jpeg_quality", "60", "042", "md_smb_quick_reply", "5183", "c", "1343", "40", "1230", "043", "044", "mms_cat_v1_forward_hot_override_enabled", "user_notice", "ptt_waveform_send", "047", "Asia/Calcutta", "250", "md_privacy_v2", "31", "29", "128", "md_messaging_enabled", "046", "crypto", "690", "045", "enc_iv", "75", "failure", "ptt_oot_playback", "AIzaSyDR5yfaG7OG8sMTUj8kfQEb8T9pN8BM6Lk", "w", "048", "2201", "web_large_files_ui", "Asia/Makassar", "812", "status_collapse_muted", "1334", "257", "2HP4dm", "049", "patches", "1290", "43cY6T", "America/Caracas", "web_sticker_maker", "campaign", "ptt_pausable_enabled", "33", "42", "attestation", "biz", "04b", "query_linked", "s", "125", "04a", "810", "availability", "1411", "responsiveness_v2_m1", "catalog_not_created", "34", "America/Santiago", "1465", "enc_p", "04d", "status_info", "04f", "key_version", "..", "04c", "04e", "md_group_notification", "1598", "1215", "web_cart_enabled", "37", "630", "1920", "2394", "-1", "vcard", "38", "elapsed", "36", "828", "peer", "pricing_category", "1245", "invalid", "stella_ios_enabled", "2687", "45", "1528", "39", "u_is_redial_audio_1104_ctrl", "1025", "1455", "58", "2524", "2603", "054", "bsp_system_message_enabled", "web_pip_redesign", "051", "verify_apps", "1974", "1272", "1322", "1755", "052", "70", "050", "1063", "1135", "1361", "80", "1096", "1828", "1851", "1251", "1921", "key_config_id", "1254", "1566", "1252", "2525", "critical_block", "1669", "max_available", "w:auth:backup:token", "product", "2530", "870", "1022", "participant_uuid", "web_cart_on_off", "1255", "1432", "1867", "41", "1415", "1440", "240", "1204", "1608", "1690", "1846", "1483", "1687", "1749", "69", "url_number", "053", "1325", "1040", "365", "59", "Asia/Riyadh", "1177", "test_recommended", "057", "1612", "43", "1061", "1518", "1635", "055", "1034", "1375", "750", "1430", "event_code", "1682", "503", "55", "865", "78", "1309", "1365", "44", "America/Guayaquil", "535", "LIMITED", "1377", "1613", "1420", "1599", "1822", "05a", "1681", "password", "1111", "1214", "1376", "1478", "47", "1082", "4282", "Europe/Istanbul", "1307", "46", "058", "1124", "256", "rate-overlimit", "retail", "u_a_socket_err_fix_succ_test", "1292", "1370", "1388", "520", "861", "psa", "regular", "1181", "1766", "05b", "1183", "1213", "1304", "1537"},
		{"1724", "profile_picture", "1071", "1314", "1605", "407", "990", "1710", "746", "pricing_model", "056", "059", "061", "1119", "6027", "65", "877", "1607", "05d", "917", "seen", "1516", "49", "470", "973", "1037", "1350", "1394", "1480", "1796", "keys", "794", "1536", "1594", "2378", "1333", "1524", "1825", "116", "309", "52", "808", "827", "909", "495", "1660", "361", "957", "google", "1357", "1565", "1967", "996", "1775", "586", "736", "1052", "1670", "bank", "177", "1416", "2194", "2222", "1454", "1839", "1275", "53", "997", "1629", "6028", "smba", "1378", "1410", "05c", "1849", "727", "create", "1559", "536", "1106", "1310", "1944", "670", "1297", "1316", "1762", "en", "1148", "1295", "1551", "1853", "1890", "1208", "1784", "7200", "05f", "178", "1283", "1332", "381", "643", "1056", "1238", "2024", "2387", "179", "981", "1547", "1705", "05e", "290", "903", "1069", "1285", "2436", "062", "251", "560", "582", "719", "56", "1700", "2321", "325", "448", "613", "777", "791", "51", "488", "902", "Asia/Almaty", "is_hidden", "1398", "1527", "1893", "1999", "2367", "2642", "237", "busy", "065", "067", "233", "590", "993", "1511", "54", "723", "860", "363", "487", "522", "605", "995", "1321", "1691", "1865", "2447", "2462", "NON_TRANSACTIONAL", "433", "871", "432", "1004", "1207", "2032", "2050", "2379", "2446", "279", "636", "703", "904", "248", "370", "691", "700", "1068", "1655", "2334", "060", "063", "364", "533", "534", "567", "1191", "1210", "1473", "1827", "069", "701", "2531", "514", "prev_dhash", "064", "496", "790", "1046", "1139", "1505", "1521", "1108", "207", "544", "637", "final", "1173", "1293", "1694", "1939", "1951", "1993", "2353", "2515", "504", "601", "857", "modify", "spam_request", "p_121_aa_1101_test4", "866", "1427", "1502", "1638", "1744", "2153", "068", "382", "725", "1704", "1864", "1990", "2003", "Asia/Dubai", "508", "531", "1387", "1474", "1632", "2307", "2386", "819", "2014", "066", "387", "1468", "1706", "2186", "2261", "471", "728", "1147", "1372", "1961"},
	}
)

// DictVersion is the version number of the token lists above.
// It's sent when connecting to the websocket so the server knows which tokens are supported.
const DictVersion = 3

type doubleByteTokenIndex struct {
	dictionary byte
	index      byte
}

var mdSingleByteTokenIndex map[string]byte
var mdDoubleByteTokenIndex map[string]doubleByteTokenIndex

func init() {
	mdSingleByteTokenIndex = make(map[string]byte, len(SingleByteTokens))
	mdDoubleByteTokenIndex = make(map[string]doubleByteTokenIndex, 256*4)
	for index, token := range SingleByteTokens {
		if token != "" {
			mdSingleByteTokenIndex[token] = byte(index)
		}
	}
	for dictIndex, tokens := range DoubleByteTokens {
		for index, token := range tokens {
			mdDoubleByteTokenIndex[token] = doubleByteTokenIndex{byte(dictIndex), byte(index)}
		}
	}
}

// GetDoubleToken gets the string value of the double-byte token at the given index.
func GetDoubleToken(index1, index2 int) (string, error) {
	if index1 < 0 || index1 >= len(DoubleByteTokens) {
		return "", fmt.Errorf("index out of double byte token bounds %d-%d", index1, index2)
	} else if index2 < 0 || index2 >= len(DoubleByteTokens[index1]) {
		return "", fmt.Errorf("index out of double byte token index %d bounds %d", index1, index2)
	}

	return DoubleByteTokens[index1][index2], nil
}

// IndexOfSingleToken gets the index of the single-byte token with the given string value.
// If there's no token with the given value, the second return value will be false.
func IndexOfSingleToken(token string) (val byte, ok bool) {
	val, ok = mdSingleByteTokenIndex[token]
	return
}

// IndexOfDoubleByteToken gets the index of the double-byte token with the given string value.
// If there's no token with the given value, the third return value will be false.
func IndexOfDoubleByteToken(token string) (byte, byte, bool) {
	val, ok := mdDoubleByteTokenIndex[token]
	return val.dictionary, val.index, ok
}

// Type tokens used in the binary XML representation.
const (
	ListEmpty   = 0
	Dictionary0 = 236
	Dictionary1 = 237
	Dictionary2 = 238
	Dictionary3 = 239
	InteropJID  = 245
	FBJID       = 246
	ADJID       = 247
	List8       = 248
	List16      = 249
	JIDPair     = 250
	Hex8        = 251
	Binary8     = 252
	Binary20    = 253
	Binary32    = 254
	Nibble8     = 255
)

// Other constants
const (
	PackedMax     = 127
	SingleByteMax = 256
)
