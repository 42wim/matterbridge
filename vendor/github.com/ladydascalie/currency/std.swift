import UIKit

struct Currency {
    /// The `ISO 4217` currency code
    var code: String

    /// The number of digits to display after the decimal point when displaying the currency
    var minorUnits: Int

    /// The factor to divide the currency figure by before handing to a currency formatter
    var factor: Int
}

class Currencies {
    static let AED: Currency = Currency(code: "AED", minorUnits: 2, factor: 100)
    static let AFN: Currency = Currency(code: "AFN", minorUnits: 2, factor: 100)
    static let ALL: Currency = Currency(code: "ALL", minorUnits: 2, factor: 100)
    static let AMD: Currency = Currency(code: "AMD", minorUnits: 2, factor: 100)
    static let ANG: Currency = Currency(code: "ANG", minorUnits: 2, factor: 100)
    static let AOA: Currency = Currency(code: "AOA", minorUnits: 2, factor: 100)
    static let ARS: Currency = Currency(code: "ARS", minorUnits: 2, factor: 100)
    static let AUD: Currency = Currency(code: "AUD", minorUnits: 2, factor: 100)
    static let AWG: Currency = Currency(code: "AWG", minorUnits: 2, factor: 100)
    static let AZN: Currency = Currency(code: "AZN", minorUnits: 2, factor: 100)
    static let BAM: Currency = Currency(code: "BAM", minorUnits: 2, factor: 100)
    static let BBD: Currency = Currency(code: "BBD", minorUnits: 2, factor: 100)
    static let BDT: Currency = Currency(code: "BDT", minorUnits: 2, factor: 100)
    static let BGN: Currency = Currency(code: "BGN", minorUnits: 2, factor: 100)
    static let BHD: Currency = Currency(code: "BHD", minorUnits: 3, factor: 1000)
    static let BIF: Currency = Currency(code: "BIF", minorUnits: 0, factor: 1)
    static let BMD: Currency = Currency(code: "BMD", minorUnits: 2, factor: 100)
    static let BND: Currency = Currency(code: "BND", minorUnits: 2, factor: 100)
    static let BOB: Currency = Currency(code: "BOB", minorUnits: 2, factor: 100)
    static let BOV: Currency = Currency(code: "BOV", minorUnits: 2, factor: 100)
    static let BRL: Currency = Currency(code: "BRL", minorUnits: 2, factor: 100)
    static let BSD: Currency = Currency(code: "BSD", minorUnits: 2, factor: 100)
    static let BTN: Currency = Currency(code: "BTN", minorUnits: 2, factor: 100)
    static let BWP: Currency = Currency(code: "BWP", minorUnits: 2, factor: 100)
    static let BYN: Currency = Currency(code: "BYN", minorUnits: 2, factor: 100)
    static let BZD: Currency = Currency(code: "BZD", minorUnits: 2, factor: 100)
    static let CAD: Currency = Currency(code: "CAD", minorUnits: 2, factor: 100)
    static let CDF: Currency = Currency(code: "CDF", minorUnits: 2, factor: 100)
    static let CHE: Currency = Currency(code: "CHE", minorUnits: 2, factor: 100)
    static let CHF: Currency = Currency(code: "CHF", minorUnits: 2, factor: 100)
    static let CHW: Currency = Currency(code: "CHW", minorUnits: 2, factor: 100)
    static let CLF: Currency = Currency(code: "CLF", minorUnits: 4, factor: 10000)
    static let CLP: Currency = Currency(code: "CLP", minorUnits: 0, factor: 1)
    static let CNY: Currency = Currency(code: "CNY", minorUnits: 2, factor: 100)
    static let COP: Currency = Currency(code: "COP", minorUnits: 2, factor: 100)
    static let COU: Currency = Currency(code: "COU", minorUnits: 2, factor: 100)
    static let CRC: Currency = Currency(code: "CRC", minorUnits: 2, factor: 100)
    static let CUC: Currency = Currency(code: "CUC", minorUnits: 2, factor: 100)
    static let CUP: Currency = Currency(code: "CUP", minorUnits: 2, factor: 100)
    static let CVE: Currency = Currency(code: "CVE", minorUnits: 2, factor: 100)
    static let CZK: Currency = Currency(code: "CZK", minorUnits: 2, factor: 100)
    static let DJF: Currency = Currency(code: "DJF", minorUnits: 0, factor: 1)
    static let DKK: Currency = Currency(code: "DKK", minorUnits: 2, factor: 100)
    static let DOP: Currency = Currency(code: "DOP", minorUnits: 2, factor: 100)
    static let DZD: Currency = Currency(code: "DZD", minorUnits: 2, factor: 100)
    static let EGP: Currency = Currency(code: "EGP", minorUnits: 2, factor: 100)
    static let ERN: Currency = Currency(code: "ERN", minorUnits: 2, factor: 100)
    static let ETB: Currency = Currency(code: "ETB", minorUnits: 2, factor: 100)
    static let EUR: Currency = Currency(code: "EUR", minorUnits: 2, factor: 100)
    static let FJD: Currency = Currency(code: "FJD", minorUnits: 2, factor: 100)
    static let FKP: Currency = Currency(code: "FKP", minorUnits: 2, factor: 100)
    static let GBP: Currency = Currency(code: "GBP", minorUnits: 2, factor: 100)
    static let GEL: Currency = Currency(code: "GEL", minorUnits: 2, factor: 100)
    static let GHS: Currency = Currency(code: "GHS", minorUnits: 2, factor: 100)
    static let GIP: Currency = Currency(code: "GIP", minorUnits: 2, factor: 100)
    static let GMD: Currency = Currency(code: "GMD", minorUnits: 2, factor: 100)
    static let GNF: Currency = Currency(code: "GNF", minorUnits: 0, factor: 1)
    static let GTQ: Currency = Currency(code: "GTQ", minorUnits: 2, factor: 100)
    static let GYD: Currency = Currency(code: "GYD", minorUnits: 2, factor: 100)
    static let HKD: Currency = Currency(code: "HKD", minorUnits: 2, factor: 100)
    static let HNL: Currency = Currency(code: "HNL", minorUnits: 2, factor: 100)
    static let HTG: Currency = Currency(code: "HTG", minorUnits: 2, factor: 100)
    static let HUF: Currency = Currency(code: "HUF", minorUnits: 2, factor: 100)
    static let IDR: Currency = Currency(code: "IDR", minorUnits: 2, factor: 100)
    static let ILS: Currency = Currency(code: "ILS", minorUnits: 2, factor: 100)
    static let INR: Currency = Currency(code: "INR", minorUnits: 2, factor: 100)
    static let IQD: Currency = Currency(code: "IQD", minorUnits: 3, factor: 1000)
    static let IRR: Currency = Currency(code: "IRR", minorUnits: 2, factor: 100)
    static let ISK: Currency = Currency(code: "ISK", minorUnits: 0, factor: 1)
    static let JMD: Currency = Currency(code: "JMD", minorUnits: 2, factor: 100)
    static let JOD: Currency = Currency(code: "JOD", minorUnits: 3, factor: 1000)
    static let JPY: Currency = Currency(code: "JPY", minorUnits: 0, factor: 1)
    static let KES: Currency = Currency(code: "KES", minorUnits: 2, factor: 100)
    static let KGS: Currency = Currency(code: "KGS", minorUnits: 2, factor: 100)
    static let KHR: Currency = Currency(code: "KHR", minorUnits: 2, factor: 100)
    static let KMF: Currency = Currency(code: "KMF", minorUnits: 0, factor: 1)
    static let KPW: Currency = Currency(code: "KPW", minorUnits: 2, factor: 100)
    static let KRW: Currency = Currency(code: "KRW", minorUnits: 0, factor: 1)
    static let KWD: Currency = Currency(code: "KWD", minorUnits: 3, factor: 1000)
    static let KYD: Currency = Currency(code: "KYD", minorUnits: 2, factor: 100)
    static let KZT: Currency = Currency(code: "KZT", minorUnits: 2, factor: 100)
    static let LAK: Currency = Currency(code: "LAK", minorUnits: 2, factor: 100)
    static let LBP: Currency = Currency(code: "LBP", minorUnits: 2, factor: 100)
    static let LKR: Currency = Currency(code: "LKR", minorUnits: 2, factor: 100)
    static let LRD: Currency = Currency(code: "LRD", minorUnits: 2, factor: 100)
    static let LSL: Currency = Currency(code: "LSL", minorUnits: 2, factor: 100)
    static let LYD: Currency = Currency(code: "LYD", minorUnits: 3, factor: 1000)
    static let MAD: Currency = Currency(code: "MAD", minorUnits: 2, factor: 100)
    static let MDL: Currency = Currency(code: "MDL", minorUnits: 2, factor: 100)
    static let MGA: Currency = Currency(code: "MGA", minorUnits: 2, factor: 100)
    static let MKD: Currency = Currency(code: "MKD", minorUnits: 2, factor: 100)
    static let MMK: Currency = Currency(code: "MMK", minorUnits: 2, factor: 100)
    static let MNT: Currency = Currency(code: "MNT", minorUnits: 2, factor: 100)
    static let MOP: Currency = Currency(code: "MOP", minorUnits: 2, factor: 100)
    static let MRU: Currency = Currency(code: "MRU", minorUnits: 2, factor: 100)
    static let MUR: Currency = Currency(code: "MUR", minorUnits: 2, factor: 100)
    static let MVR: Currency = Currency(code: "MVR", minorUnits: 2, factor: 100)
    static let MWK: Currency = Currency(code: "MWK", minorUnits: 2, factor: 100)
    static let MXN: Currency = Currency(code: "MXN", minorUnits: 2, factor: 100)
    static let MXV: Currency = Currency(code: "MXV", minorUnits: 2, factor: 100)
    static let MYR: Currency = Currency(code: "MYR", minorUnits: 2, factor: 100)
    static let MZN: Currency = Currency(code: "MZN", minorUnits: 2, factor: 100)
    static let NAD: Currency = Currency(code: "NAD", minorUnits: 2, factor: 100)
    static let NGN: Currency = Currency(code: "NGN", minorUnits: 2, factor: 100)
    static let NIO: Currency = Currency(code: "NIO", minorUnits: 2, factor: 100)
    static let NOK: Currency = Currency(code: "NOK", minorUnits: 2, factor: 100)
    static let NPR: Currency = Currency(code: "NPR", minorUnits: 2, factor: 100)
    static let NZD: Currency = Currency(code: "NZD", minorUnits: 2, factor: 100)
    static let OMR: Currency = Currency(code: "OMR", minorUnits: 3, factor: 1000)
    static let PAB: Currency = Currency(code: "PAB", minorUnits: 2, factor: 100)
    static let PEN: Currency = Currency(code: "PEN", minorUnits: 2, factor: 100)
    static let PGK: Currency = Currency(code: "PGK", minorUnits: 2, factor: 100)
    static let PHP: Currency = Currency(code: "PHP", minorUnits: 2, factor: 100)
    static let PKR: Currency = Currency(code: "PKR", minorUnits: 2, factor: 100)
    static let PLN: Currency = Currency(code: "PLN", minorUnits: 2, factor: 100)
    static let PYG: Currency = Currency(code: "PYG", minorUnits: 0, factor: 1)
    static let QAR: Currency = Currency(code: "QAR", minorUnits: 2, factor: 100)
    static let RON: Currency = Currency(code: "RON", minorUnits: 2, factor: 100)
    static let RSD: Currency = Currency(code: "RSD", minorUnits: 2, factor: 100)
    static let RUB: Currency = Currency(code: "RUB", minorUnits: 2, factor: 100)
    static let RWF: Currency = Currency(code: "RWF", minorUnits: 0, factor: 1)
    static let SAR: Currency = Currency(code: "SAR", minorUnits: 2, factor: 100)
    static let SBD: Currency = Currency(code: "SBD", minorUnits: 2, factor: 100)
    static let SCR: Currency = Currency(code: "SCR", minorUnits: 2, factor: 100)
    static let SDG: Currency = Currency(code: "SDG", minorUnits: 2, factor: 100)
    static let SEK: Currency = Currency(code: "SEK", minorUnits: 2, factor: 100)
    static let SGD: Currency = Currency(code: "SGD", minorUnits: 2, factor: 100)
    static let SHP: Currency = Currency(code: "SHP", minorUnits: 2, factor: 100)
    static let SLE: Currency = Currency(code: "SLE", minorUnits: 2, factor: 100)
    static let SLL: Currency = Currency(code: "SLL", minorUnits: 2, factor: 100)
    static let SOS: Currency = Currency(code: "SOS", minorUnits: 2, factor: 100)
    static let SRD: Currency = Currency(code: "SRD", minorUnits: 2, factor: 100)
    static let SSP: Currency = Currency(code: "SSP", minorUnits: 2, factor: 100)
    static let STN: Currency = Currency(code: "STN", minorUnits: 2, factor: 100)
    static let SVC: Currency = Currency(code: "SVC", minorUnits: 2, factor: 100)
    static let SYP: Currency = Currency(code: "SYP", minorUnits: 2, factor: 100)
    static let SZL: Currency = Currency(code: "SZL", minorUnits: 2, factor: 100)
    static let THB: Currency = Currency(code: "THB", minorUnits: 2, factor: 100)
    static let TJS: Currency = Currency(code: "TJS", minorUnits: 2, factor: 100)
    static let TMT: Currency = Currency(code: "TMT", minorUnits: 2, factor: 100)
    static let TND: Currency = Currency(code: "TND", minorUnits: 3, factor: 1000)
    static let TOP: Currency = Currency(code: "TOP", minorUnits: 2, factor: 100)
    static let TRY: Currency = Currency(code: "TRY", minorUnits: 2, factor: 100)
    static let TTD: Currency = Currency(code: "TTD", minorUnits: 2, factor: 100)
    static let TWD: Currency = Currency(code: "TWD", minorUnits: 2, factor: 100)
    static let TZS: Currency = Currency(code: "TZS", minorUnits: 2, factor: 100)
    static let UAH: Currency = Currency(code: "UAH", minorUnits: 2, factor: 100)
    static let UGX: Currency = Currency(code: "UGX", minorUnits: 0, factor: 1)
    static let USD: Currency = Currency(code: "USD", minorUnits: 2, factor: 100)
    static let USN: Currency = Currency(code: "USN", minorUnits: 2, factor: 100)
    static let UYI: Currency = Currency(code: "UYI", minorUnits: 0, factor: 1)
    static let UYU: Currency = Currency(code: "UYU", minorUnits: 2, factor: 100)
    static let UYW: Currency = Currency(code: "UYW", minorUnits: 4, factor: 10000)
    static let UZS: Currency = Currency(code: "UZS", minorUnits: 2, factor: 100)
    static let VED: Currency = Currency(code: "VED", minorUnits: 2, factor: 100)
    static let VES: Currency = Currency(code: "VES", minorUnits: 2, factor: 100)
    static let VND: Currency = Currency(code: "VND", minorUnits: 0, factor: 1)
    static let VUV: Currency = Currency(code: "VUV", minorUnits: 0, factor: 1)
    static let WST: Currency = Currency(code: "WST", minorUnits: 2, factor: 100)
    static let XAF: Currency = Currency(code: "XAF", minorUnits: 0, factor: 1)
    static let XAG: Currency = Currency(code: "XAG", minorUnits: 0, factor: 1)
    static let XAU: Currency = Currency(code: "XAU", minorUnits: 0, factor: 1)
    static let XBA: Currency = Currency(code: "XBA", minorUnits: 0, factor: 1)
    static let XBB: Currency = Currency(code: "XBB", minorUnits: 0, factor: 1)
    static let XBC: Currency = Currency(code: "XBC", minorUnits: 0, factor: 1)
    static let XBD: Currency = Currency(code: "XBD", minorUnits: 0, factor: 1)
    static let XCD: Currency = Currency(code: "XCD", minorUnits: 2, factor: 100)
    static let XDR: Currency = Currency(code: "XDR", minorUnits: 0, factor: 1)
    static let XOF: Currency = Currency(code: "XOF", minorUnits: 0, factor: 1)
    static let XPD: Currency = Currency(code: "XPD", minorUnits: 0, factor: 1)
    static let XPF: Currency = Currency(code: "XPF", minorUnits: 0, factor: 1)
    static let XPT: Currency = Currency(code: "XPT", minorUnits: 0, factor: 1)
    static let XSU: Currency = Currency(code: "XSU", minorUnits: 0, factor: 1)
    static let XTS: Currency = Currency(code: "XTS", minorUnits: 0, factor: 1)
    static let XUA: Currency = Currency(code: "XUA", minorUnits: 0, factor: 1)
    static let XXX: Currency = Currency(code: "XXX", minorUnits: 0, factor: 1)
    static let YER: Currency = Currency(code: "YER", minorUnits: 2, factor: 100)
    static let ZAR: Currency = Currency(code: "ZAR", minorUnits: 2, factor: 100)
    static let ZMW: Currency = Currency(code: "ZMW", minorUnits: 2, factor: 100)
    static let ZWL: Currency = Currency(code: "ZWL", minorUnits: 2, factor: 100)
    

    static let allCurrencies: [String: Currency] =
        [
            "AED": AED,
            "AFN": AFN,
            "ALL": ALL,
            "AMD": AMD,
            "ANG": ANG,
            "AOA": AOA,
            "ARS": ARS,
            "AUD": AUD,
            "AWG": AWG,
            "AZN": AZN,
            "BAM": BAM,
            "BBD": BBD,
            "BDT": BDT,
            "BGN": BGN,
            "BHD": BHD,
            "BIF": BIF,
            "BMD": BMD,
            "BND": BND,
            "BOB": BOB,
            "BOV": BOV,
            "BRL": BRL,
            "BSD": BSD,
            "BTN": BTN,
            "BWP": BWP,
            "BYN": BYN,
            "BZD": BZD,
            "CAD": CAD,
            "CDF": CDF,
            "CHE": CHE,
            "CHF": CHF,
            "CHW": CHW,
            "CLF": CLF,
            "CLP": CLP,
            "CNY": CNY,
            "COP": COP,
            "COU": COU,
            "CRC": CRC,
            "CUC": CUC,
            "CUP": CUP,
            "CVE": CVE,
            "CZK": CZK,
            "DJF": DJF,
            "DKK": DKK,
            "DOP": DOP,
            "DZD": DZD,
            "EGP": EGP,
            "ERN": ERN,
            "ETB": ETB,
            "EUR": EUR,
            "FJD": FJD,
            "FKP": FKP,
            "GBP": GBP,
            "GEL": GEL,
            "GHS": GHS,
            "GIP": GIP,
            "GMD": GMD,
            "GNF": GNF,
            "GTQ": GTQ,
            "GYD": GYD,
            "HKD": HKD,
            "HNL": HNL,
            "HTG": HTG,
            "HUF": HUF,
            "IDR": IDR,
            "ILS": ILS,
            "INR": INR,
            "IQD": IQD,
            "IRR": IRR,
            "ISK": ISK,
            "JMD": JMD,
            "JOD": JOD,
            "JPY": JPY,
            "KES": KES,
            "KGS": KGS,
            "KHR": KHR,
            "KMF": KMF,
            "KPW": KPW,
            "KRW": KRW,
            "KWD": KWD,
            "KYD": KYD,
            "KZT": KZT,
            "LAK": LAK,
            "LBP": LBP,
            "LKR": LKR,
            "LRD": LRD,
            "LSL": LSL,
            "LYD": LYD,
            "MAD": MAD,
            "MDL": MDL,
            "MGA": MGA,
            "MKD": MKD,
            "MMK": MMK,
            "MNT": MNT,
            "MOP": MOP,
            "MRU": MRU,
            "MUR": MUR,
            "MVR": MVR,
            "MWK": MWK,
            "MXN": MXN,
            "MXV": MXV,
            "MYR": MYR,
            "MZN": MZN,
            "NAD": NAD,
            "NGN": NGN,
            "NIO": NIO,
            "NOK": NOK,
            "NPR": NPR,
            "NZD": NZD,
            "OMR": OMR,
            "PAB": PAB,
            "PEN": PEN,
            "PGK": PGK,
            "PHP": PHP,
            "PKR": PKR,
            "PLN": PLN,
            "PYG": PYG,
            "QAR": QAR,
            "RON": RON,
            "RSD": RSD,
            "RUB": RUB,
            "RWF": RWF,
            "SAR": SAR,
            "SBD": SBD,
            "SCR": SCR,
            "SDG": SDG,
            "SEK": SEK,
            "SGD": SGD,
            "SHP": SHP,
            "SLE": SLE,
            "SLL": SLL,
            "SOS": SOS,
            "SRD": SRD,
            "SSP": SSP,
            "STN": STN,
            "SVC": SVC,
            "SYP": SYP,
            "SZL": SZL,
            "THB": THB,
            "TJS": TJS,
            "TMT": TMT,
            "TND": TND,
            "TOP": TOP,
            "TRY": TRY,
            "TTD": TTD,
            "TWD": TWD,
            "TZS": TZS,
            "UAH": UAH,
            "UGX": UGX,
            "USD": USD,
            "USN": USN,
            "UYI": UYI,
            "UYU": UYU,
            "UYW": UYW,
            "UZS": UZS,
            "VED": VED,
            "VES": VES,
            "VND": VND,
            "VUV": VUV,
            "WST": WST,
            "XAF": XAF,
            "XAG": XAG,
            "XAU": XAU,
            "XBA": XBA,
            "XBB": XBB,
            "XBC": XBC,
            "XBD": XBD,
            "XCD": XCD,
            "XDR": XDR,
            "XOF": XOF,
            "XPD": XPD,
            "XPF": XPF,
            "XPT": XPT,
            "XSU": XSU,
            "XTS": XTS,
            "XUA": XUA,
            "XXX": XXX,
            "YER": YER,
            "ZAR": ZAR,
            "ZMW": ZMW,
            "ZWL": ZWL
            ]

    /// Look up a currency struct containing currency and formatting information
    ///
    /// - Parameter code: The `ISO 4217` currency code to search for
    /// - Returns: A `Currency` object, if supported. Otherwise nil.
    class func currency(for code: String) -> Currency? {
        return allCurrencies[code]
    }
}
