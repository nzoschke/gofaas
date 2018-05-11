# Intro to Go Functions-as-a-Service and Lambda
### Why FaaS matters and why Go is well-suited for Lambda

## Intro to FaaS and Serverless

Functions-as-a-Service (FaaS) are one of the latest advances in cloud Infrastructure-as-a-Service (IaaS). FaaS fits neatly under the "serverless" label -- managed services that shield us users from any details about the underlying servers.

Serverless isn't really new. S3 is one of the oldest cloud services, and has seen nearly universal adoption because us users can upload tons of data and let AWS worry about the computers, networks, hard drives and software required to never lose our data. S3 fits the definition of "serverless" to a tee. API Gateway and DynamoDB are other key serverless services.

But Lambda is fairly new and does represent a shift in computing. Before Lambda, us users have been responsible for provisioning servers, instances or VMs to run our software. We have also been responsible for designing architectures that are resilient to instance failures, and that can scale instances horizontally and/or vertically as demand increases. Lambda promises to remove these worries. We upload a .zip file of our function to S3, and AWS manages all the infrastructure to run our function.

Lambda also represents a shift in programming to an event-driven architecture. Our code is no longer running on a server 24/7 listening for requests. Instead it lays dormant until AWS invokes it with an event. An obvious event is an HTTP request, delivered to our function by API Gateway. But AWS offers many other interesting event sources, like when a S3 bucket receives a new file, or a pre-configured schedule.

## Why is it Important?

Lambda and other serverless services promise a lot to us users.

Serverless **reduces operational overhead**. We are responsible for our function code and data, and AWS is responsible for everything else -- event delivery, function execution, storage, replication and all the underlying infrastructure. It's now impossible to to get paged for a crashed HTTP server, or an out-of-capacity container cluster.

Lambda **increases simplicity**. We only need to reason about events in and events out. Whole classes of long-running software -- like HTTP servers, worker dispatch queues, or S3 bucket pollers -- are eliminated. We can focus entirely on our business logic.

Lambda **increases security**. Every function gets its own environment variables and execution policy. We no longer have to give all the keys to our entire application. An auth function can be the only function that knows the API password or OAuth secret. A single API endpoint can be the only function that can write data to S3.

Lambda can **reduce cost**. On the low end, functions cost nothing when idle, and fractions of a penny per 100 milliseconds when running. It's common to see bills of a few cents for simple apps. There are tough questions about the cost of Lambda at huge scale. But consider a new possibility: with a FaaS app, you can drill into your bill to see what functions cost the most, and optimize those. Optimizing utilization on a production container cluster is extremely difficult.

There are still tradeoffs and challenges with FaaS. But just like how you would be considered crazy now to manage your own FTP server, someday you'll be crazy to reach for an instance or container where a function will do.

## Why Go?

With Lambda we focus entirely on writing functions, and offers a few choices for programming languages: C#, Go, Python, Java and JavaScript. Each language has its strong suit, but some features of Go make it particularly well-suited for Lambda.

Go's release management offers **stability**. Lambda supports 1.x family of Go, which means code that is 7 years old and code that will be written a year from now will all work. On the opposite end of the spectrum we have Javascript, where AWS only offers a 2 year old version of the Node.JS runtime.

Go's type system and error handling offers **correctness**. Our functions take events in and return events (or an error).

Go's cross-compiler solves **packaging**. Every laptop with the Go tool chain can build Linux binaries ready for Lambda. We don't need Docker or Linux build services to produce a suitable Lambda package.

Go's binary program format offers **speed**. Because Lambda runs new versions of our program on-demand, slow boot times can turn into a real problem. Go's binaries have very little overhead to start, compared to a dynamic VM like Python or Java.

Go's context pattern offers **observability**. Google sponsored the development of Go with a big goal in mind - to make large scale distributed systems programming easier. Thanks to the guidance and expertise of the Googler's behind Go, it comes with important observability primitives out of the box.

## Summary

Lambda and Go offer many features that make them great choices for building application logic.

- Write functions
- Cross-compile binaries
- Configure Lambda functions
- Configure other serverless infrastructure and events

And that's about it.

Many responsibilities are no longer our problems:

- Building AMIs
- Running, scaling and failing over instances
- Managing container clusters
- Building container images

We get to worry about less, and likely pay less. No wonder there's so much hype around serverless, FaaS and Lambda. And it seems like Go is a great choice for writing all those functions.
