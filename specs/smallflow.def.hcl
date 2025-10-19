system "smallflow" {
  version        = "1.0.0"
  description    = "Build, run, and observe workflows without the overhead!"
  repository_url = "https://github.com/morebec/smallflow"

  implementation {
    backend {
      language = "go"
      version  = 1.24
    }
  }

  non_functional_requirement "GlobalPerformance" {
    category    = "Performance"
    description = "The system should provide quick response times under load."
  }
  non_functional_requirement "DataSecurity" {
    category    = "Security"
    description = "The system must ensure that all data is handled securely."
  }
  service_level_agreement "APIResponseTime" {
    description = "The system should respond to 95% of requests within 200ms."
    metric      = "Response Time"
    target      = "200ms"
  }
  service_level_agreement "SystemUptime" {
    description = "The system should have an uptime of 99.9% or higher."
    metric      = "Uptime"
    target      = "99.9%"
  }
}
