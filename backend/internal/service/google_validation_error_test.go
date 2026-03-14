package service

import "testing"

func TestExtractGoogleValidationRequired(t *testing.T) {
	body := []byte(`{
	  "error": {
	    "code": 403,
	    "message": "Verify your account to continue.",
	    "status": "PERMISSION_DENIED",
	    "details": [
	      {
	        "@type": "type.googleapis.com/google.rpc.ErrorInfo",
	        "reason": "VALIDATION_REQUIRED",
	        "domain": "cloudcode-pa.googleapis.com",
	        "metadata": {
	          "validation_error_message": "Verify your account to continue.",
	          "validation_url_link_text": "Verify your account",
	          "validation_url": "https://accounts.google.com/signin/continue?foo=bar",
	          "validation_learn_more_link_text": "Learn more",
	          "validation_learn_more_url": "https://support.google.com/accounts?p=al_alert"
	        }
	      },
	      {
	        "@type": "type.googleapis.com/google.rpc.Help",
	        "links": [
	          {
	            "description": "Verify your account",
	            "url": "https://accounts.google.com/signin/continue?foo=bar"
	          },
	          {
	            "description": "Learn more",
	            "url": "https://support.google.com/accounts?p=al_alert"
	          }
	        ]
	      }
	    ]
	  }
	}`)

	info := ExtractGoogleValidationRequired(body)
	if info == nil {
		t.Fatal("expected validation info")
	}
	if got, want := info.ValidationURL, "https://accounts.google.com/signin/continue?foo=bar"; got != want {
		t.Fatalf("validation url = %q, want %q", got, want)
	}
	if got, want := info.ValidationLabel, "Verify your account"; got != want {
		t.Fatalf("validation label = %q, want %q", got, want)
	}
	if got, want := info.LearnMoreURL, "https://support.google.com/accounts?p=al_alert"; got != want {
		t.Fatalf("learn more url = %q, want %q", got, want)
	}
	if got, want := info.Reason, "VALIDATION_REQUIRED"; got != want {
		t.Fatalf("reason = %q, want %q", got, want)
	}
}

func TestExtractGoogleValidationRequired_IgnoresNonHTTPS(t *testing.T) {
	body := []byte(`{
	  "error": {
	    "message": "Verify your account to continue.",
	    "details": [
	      {
	        "@type": "type.googleapis.com/google.rpc.ErrorInfo",
	        "reason": "VALIDATION_REQUIRED",
	        "domain": "cloudcode-pa.googleapis.com",
	        "metadata": {
	          "validation_url": "javascript:alert(1)"
	        }
	      }
	    ]
	  }
	}`)

	if info := ExtractGoogleValidationRequired(body); info != nil {
		t.Fatal("expected nil for invalid validation url")
	}
}

func TestExtractGoogleValidationRequiredFromText(t *testing.T) {
	text := `API 返回 403: {"error":{"code":403,"message":"Verify your account to continue.","details":[{"@type":"type.googleapis.com/google.rpc.ErrorInfo","reason":"VALIDATION_REQUIRED","metadata":{"validation_url":"https://accounts.google.com/signin/continue?foo=bar","validation_url_link_text":"Verify your account","validation_learn_more_url":"https://support.google.com/accounts?p=al_alert","validation_learn_more_link_text":"Learn more","validation_error_message":"Verify your account to continue."}}]}}`

	info := ExtractGoogleValidationRequiredFromText(text)
	if info == nil {
		t.Fatal("expected validation info")
	}
	if got, want := info.ValidationURL, "https://accounts.google.com/signin/continue?foo=bar"; got != want {
		t.Fatalf("validation url = %q, want %q", got, want)
	}
	if got, want := info.ValidationLabel, "Verify your account"; got != want {
		t.Fatalf("validation label = %q, want %q", got, want)
	}
}
