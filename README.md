# function-sdk-go

The Go SDK for Composition Functions.

This SDK is currently a beta and does not yet have a stable API. It follows the
same [contributing guidelines] as Crossplane.

```go
func (f *Function) RunFunction(_ context.Context, req *fnv1beta1.RunFunctionRequest) (*fnv1beta1.RunFunctionResponse, error) {
	// This creates a new response to the supplied request. Note that Functions
	// are run in a pipeline! Other Functions may have run before this one. If
	// they did, response.To will copy their desired state from req to rsp. Be
	// sure to pass through any desired state your Function is not concerned
	// with unmodified.
	rsp := response.To(req, response.DefaultTTL)

	// Input is supplied by the author of a Composition when they choose to run
	// your Function. Input is arbitrary, except that it must be a KRM-like
	// object. Supporting input is also optional - if you don't need to you can
	// delete this, and delete the input directory.
	in := &v1beta1.Input{}
	if err := request.GetInput(req, in); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get Function input from %T", req))
		return rsp, nil
	}

	// Get the observed composite resource (XR) from the request. There should
	// always be an observed XR in the request - this represents the current
	// state of the XR.
	oxr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed XR from %T", req))
		return rsp, nil
	}

	// Read the desired number of widgets from our observed XR. We don't have
	// a struct for the XR, so we use an unstructured, fieldpath based getter.
	widgets, err := oxr.Resource.GetInteger("spec.widgets")
	oxr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get desired spec.widgets from observed XR"))
		return rsp, nil
	}

	// Get any existing desired composed resources from the request.
	// Desired composed resources would exist if a previous Function in the
	// pipeline added them.
	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get desired composed resources from %T", req))
		return rsp, nil
	}

	// Create a desired composed resource using unstructured data.
	desired["new"] = &resource.DesiredComposed{Resource: composed.New()}
	desired["new"].Resource.SetAPIVersion("example.org/v1")
	desired["new"].Resource.SetKind("CoolResource")

	// Set the desired composed resource's widgets to the value extracted from
	// the observed XR.
	desired["new"].Resource.SetInteger("spec.widgets", widgets)

	// You could create a desired composed resource using structured data, too.
	// db, _ := composed.From(&v1beta1.Instance{})
	// desired["database"] = &resource.DesiredComposed{Resource: db}

	// Set the labels of all desired resources, including our new one.
	for _, r := range desired {
		r.Resource.SetLabels(map[string]string{"coolness": "high"})
	}

	// Set our updated desired composed resource in the response we'll return.
	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot set desired composed resources in %T", rsp))
		return rsp, nil
	}

	return rsp, nil
}
```

[contributing guidelines]: https://github.com/crossplane/crossplane/tree/master/contributing
