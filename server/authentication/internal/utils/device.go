package utils

import "strings"

// ParseDeviceInfo takes a raw User-Agent string and returns a clean "Browser on OS" string.
func ParseDeviceInfo(userAgent string) string {
    if userAgent == "" {
        return "Unknown Device"
    }

    // 1. OS Parse karo
    os := "Unknown OS"
    switch {
    case strings.Contains(userAgent, "Windows"):
        os = "Windows"
    case strings.Contains(userAgent, "Mac OS X") || strings.Contains(userAgent, "Macintosh"):
        os = "macOS"
    case strings.Contains(userAgent, "Linux"):
        os = "Linux"
    case strings.Contains(userAgent, "Android"):
        os = "Android"
    case strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "iPad"):
        os = "iOS"
    }

    // 2. Browser Parse karo
    browser := "Unknown Browser"
    switch {
    case strings.Contains(userAgent, "Edg"):
        browser = "Edge"
    case strings.Contains(userAgent, "Chrome"): // Firefox mein bhi kabhi kabhi chrome tag hota hai par exact match kaam karega
        browser = "Chrome"
    case strings.Contains(userAgent, "Firefox"):
        browser = "Firefox"
    case strings.Contains(userAgent, "Safari") && !strings.Contains(userAgent, "Chrome"):
        browser = "Safari"
    }

    return browser + " on " + os
}