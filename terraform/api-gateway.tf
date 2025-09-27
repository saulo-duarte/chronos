resource "aws_apigatewayv2_api" "lambda_api_v2" {
  name          = "${var.lambda_function_name}-http-api"
  protocol_type = "HTTP"

  cors_configuration {
    allow_origins     = [data.aws_ssm_parameter.frontend_url.value]
    allow_methods     = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers     = ["Content-Type", "Authorization", "Cookie"]
    allow_credentials = true
    max_age           = 3600
  }
}

resource "aws_apigatewayv2_integration" "lambda_integration_v2" {
  api_id                 = aws_apigatewayv2_api.lambda_api_v2.id
  integration_type       = "AWS_PROXY"
  integration_uri        = aws_lambda_function.go_lambda.arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "lambda_route_proxy" {
  api_id    = aws_apigatewayv2_api.lambda_api_v2.id
  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration_v2.id}"
}

resource "aws_apigatewayv2_stage" "lambda_stage_v2" {
  api_id      = aws_apigatewayv2_api.lambda_api_v2.id
  name        = "$default"
  auto_deploy = true
}

resource "aws_lambda_permission" "api_gateway_v2" {
  statement_id  = "AllowExecutionFromAPIGatewayV2"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.go_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.lambda_api_v2.execution_arn}/*/*"
}
