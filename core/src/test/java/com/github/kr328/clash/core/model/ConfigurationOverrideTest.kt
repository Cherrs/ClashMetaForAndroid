package com.github.kr328.clash.core.model

import kotlinx.serialization.decodeFromString
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import kotlin.test.Test
import kotlin.test.assertContains
import kotlin.test.assertEquals
import kotlin.test.assertNotSame

class ConfigurationOverrideTest {
    private val json = Json {
        encodeDefaults = false
    }

    @Test
    fun customRulesSurviveJsonRoundTrip() {
        val configuration = ConfigurationOverride(
            prependRules = listOf(
                "DOMAIN,example.com,DIRECT",
                "MATCH,Proxy",
            ),
        )

        val encoded = json.encodeToString(configuration)

        assertContains(
            encoded,
            "\"prepend-rules\":[\"DOMAIN,example.com,DIRECT\",\"MATCH,Proxy\"]",
        )
        assertEquals(configuration, json.decodeFromString(encoded))
    }

    @Test
    fun deepCopyDoesNotShareMutableState() {
        val configuration = ConfigurationOverride(
            authentication = mutableListOf("user:password"),
            hosts = mutableMapOf("example.com" to "127.0.0.1"),
            prependRules = mutableListOf("DOMAIN,example.com,DIRECT"),
        ).apply {
            dns.nameServer = mutableListOf("1.1.1.1")
            dns.fallbackFilter.domain = mutableListOf("example.com")
            sniffer.sniff.http.ports = mutableListOf("80")
        }
        val copied = configuration.deepCopy()
        assertEquals(configuration, copied)

        (configuration.authentication as MutableList<String>).add("other:password")
        (configuration.hosts as MutableMap<String, String>)["other.example"] = "127.0.0.2"
        (configuration.prependRules as MutableList<String>).add("MATCH,DIRECT")
        (configuration.dns.nameServer as MutableList<String>).add("8.8.8.8")
        (configuration.dns.fallbackFilter.domain as MutableList<String>).add("other.example")
        (configuration.sniffer.sniff.http.ports as MutableList<String>).add("8080")
        configuration.dns.enable = true

        assertEquals(listOf("user:password"), copied.authentication)
        assertEquals(mapOf("example.com" to "127.0.0.1"), copied.hosts)
        assertEquals(listOf("DOMAIN,example.com,DIRECT"), copied.prependRules)
        assertEquals(listOf("1.1.1.1"), copied.dns.nameServer)
        assertEquals(listOf("example.com"), copied.dns.fallbackFilter.domain)
        assertEquals(listOf("80"), copied.sniffer.sniff.http.ports)
        assertEquals(null, copied.dns.enable)
        assertNotSame(configuration.dns, copied.dns)
        assertNotSame(configuration.sniffer.sniff.http, copied.sniffer.sniff.http)
    }
}
