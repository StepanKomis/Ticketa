window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "/docs/openapi.yaml",
    dom_id: '#swagger-ui',
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    layout: "StandaloneLayout",
    deepLinking: true,
    persistAuthorization: true,
  });
};
