# Static Websites
### With S3, CloudFront and ACM

A modern application pattern is to deploy static web content, say a React or Vue.js single page application, that interacts with an app API. For this example we want to deploy a static web site -- `https://www.gofaas.net` -- that talks to our gofaas API running at `https://api.gofaas.net` to authenticate a user and display a dashboard.

This pattern offers many advantages. Now the API is only concerned with data, making it easier to write and more cost effective to run. The web content is completely static, making it extremely reliable and cost effective to deliver to our users.

Compare this to a traditional Model View Controller (MVC) approach like Rails or Django. In this architecture the API may spend lots of time rendering HTML, and the HTML may not get served to users if there is an application bug or a database outage.

Static websites are a solved problem on AWS. We simply create an S3 bucket configured for website hosting and upload the content with public-read permissions. Then anyone can access the content from a URL like `http://gofaas-webbucket-572007530218.s3-website-us-east-1.amazonaws.com` with some of the highest reliability and lowest storage and bandwidth costs possible.

Serving this from a custom domain is also a solved problem. We add the CloudFront CDN, configured with an SSL cert via the AWS Certificate Manager, in front of the S3 bucket. When we point our custom domain DNS to CloudFront, users can access the content from a URL like `https://www.gofaas.net` with some of the fastest delivery times and lowest bandwidth costs possible thanks to the global content caching network.

Let's set this all up for our app...

## AWS Config

There are a lot of configuration options for S3 and CloudFront so the template isn't short. But rest assured this template will keep your website running forever.

Note that we add a `WebsiteConfiguration` for the S3 bucket, and conditionally create an ACM cert and CloudFront distribution if we specify the `WebDomainName` parameter.

```yaml
---
AWSTemplateFormatVersion: '2010-09-09'

Conditions:
  WebDomainNameSpecified: !Not [!Equals [!Ref WebDomainName, ""]]

Mappings:
  RegionMap:
    ap-northeast-1:
      S3HostedZoneId: Z2M4EHUR26P7ZW
      S3WebsiteEndpoint: s3-website-ap-northeast-1.amazonaws.com
    ap-southeast-1:
      S3HostedZoneId: Z3O0J2DXBE1FTB
      S3WebsiteEndpoint: s3-website-ap-southeast-1.amazonaws.com
    ap-southeast-2:
      S3HostedZoneId: Z1WCIGYICN2BYD
      S3WebsiteEndpoint: s3-website-ap-southeast-2.amazonaws.com
    eu-west-1:
      S3HostedZoneId: Z1BKCTXD74EZPE
      S3WebsiteEndpoint: s3-website-eu-west-1.amazonaws.com
    sa-east-1:
      S3HostedZoneId: Z31GFT0UA1I2HV
      S3WebsiteEndpoint: s3-website-sa-east-1.amazonaws.com
    us-east-1:
      S3HostedZoneId: Z3AQBSTGFYJSTF
      S3WebsiteEndpoint: s3-website-us-east-1.amazonaws.com
    us-west-1:
      S3HostedZoneId: Z2F56UZL2M1ACD
      S3WebsiteEndpoint: s3-website-us-west-1.amazonaws.com
    us-west-2:
      S3HostedZoneId: Z3BJ6K6RIION7M
      S3WebsiteEndpoint: s3-website-us-west-2.amazonaws.com

Outputs:
  WebDistributionDomainName:
    Condition: WebDomainNameSpecified
    Value: !GetAtt WebDistribution.DomainName

  WebUrl:
    Value:
      !If
      - WebDomainNameSpecified
      - !Sub https://${WebDomainName}
      - !Sub
        - http://${WebBucket}.${Endpoint}
        - {Endpoint: !FindInMap [RegionMap, !Ref "AWS::Region", S3WebsiteEndpoint]}

Parameters:
  WebDomainName:
    Default: ""
    Description: "Domain or subdomain for the static website distribution, e.g. www.gofaas.net"
    Type: String

Resources:
  WebBucket:
    DeletionPolicy: Retain
    Properties:
      AccessControl: PublicRead
      BucketName: !If [WebDomainNameSpecified, !Ref WebDomainName, !Sub "${AWS::StackName}-webbucket-${AWS::AccountId}"]
      WebsiteConfiguration:
        ErrorDocument: 404.html
        IndexDocument: index.html
    Type: AWS::S3::Bucket

  WebBucketPolicy:
    Properties:
      Bucket: !Ref WebBucket
      PolicyDocument:
        Statement:
          - Action: s3:GetObject
            Effect: Allow
            Principal: "*"
            Resource: !Sub arn:aws:s3:::${WebBucket}/*
            Sid: PublicReadForGetBucketObjects
    Type: AWS::S3::BucketPolicy

  WebCertificate:
    Condition: WebDomainNameSpecified
    Properties:
      DomainName: !Ref WebDomainName
    Type: AWS::CertificateManager::Certificate

  WebDistribution:
    Condition: WebDomainNameSpecified
    Properties:
      DistributionConfig:
        Aliases:
          - !Ref WebDomainName
        Comment: !Sub Distribution for ${WebBucket}
        DefaultCacheBehavior:
          AllowedMethods:
            - GET
            - HEAD
          Compress: true
          ForwardedValues:
            Cookies:
              Forward: none
            QueryString: true
          TargetOriginId: !Ref WebBucket
          ViewerProtocolPolicy: redirect-to-https
        DefaultRootObject: index.html
        Enabled: true
        HttpVersion: http2
        Origins:
          - CustomOriginConfig:
              HTTPPort: 80
              HTTPSPort: 443
              OriginProtocolPolicy: http-only
            DomainName: !Sub
              - ${WebBucket}.${Endpoint}
              - {Endpoint: !FindInMap [RegionMap, !Ref "AWS::Region", S3WebsiteEndpoint]}
            Id: !Ref WebBucket
        PriceClass: PriceClass_All
        ViewerCertificate:
          AcmCertificateArn: !Ref WebCertificate
          SslSupportMethod: sni-only
    Type: AWS::CloudFront::Distribution
```
> From [template.yml](../template.yml)

## Deploy

Now we can deploy the config to create the website bucket:

```shell
$ aws cloudformation package \
    --output-template-file out.yml --template-file template.yml

$ aws cloudformation deploy --stack-name gofaas \
    --capabilities CAPABILITY_NAMED_IAM --template-file out.yml
Waiting for stack create/update to complete

$ aws cloudformation describe-stacks --stack-name gofaas \
    --output text --query 'Stacks[*].Outputs'
WebUrl	http://gofaas-webbucket-572007530218.s3-website-us-east-1.amazonaws.com
```
> From [Makefile](../Makefile)

And upload our first content:

```shell
$ aws s3 sync public s3://gofaas-webbucket-572007530218/
upload: public/index.html to s3://gofaas-webbucket-572007530218/index.html
...
```
> From [Makefile](../Makefile)

Sure enough we can access it over HTTP:

```shell
$ curl http://gofaas-webbucket-572007530218.s3-website-us-east-1.amazonaws.com/
...
<title>My first gofaas/Vue app</title>
```

## Deploy Custom Domain

Now we can re-deploy the config with our domain name to create the certificate and CDN:

```shell
aws cloudformation deploy  --stack-name gofaas         \
    --parameter-overrides WebDomainName=www.gofaas.net \
    --capabilities CAPABILITY_NAMED_IAM --template-file out.yml

$ aws cloudformation describe-stacks --stack-name gofaas \
    --output text --query 'Stacks[*].Outputs'
WebDistributionDomainName  d2bwnae7bzw1t6.cloudfront.net
WebUrl                     https://www.gofaas.net
```

Note that this can take 10 to 20 minutes to set up the global infrastructure for our static site. Also note that ACM will send an email to the domain owner (e.g. admin@gofaas.net) who must click through the approval to create the certificate. See the [ACM email validation](https://docs.aws.amazon.com/acm/latest/userguide/gs-acm-validate-email.html) guide for more information.

Once the CDN is in place, we can sync content to the S3 bucket the same way, but we may need to invalidate content cached in the CDN to immediately see the latest content:

```console
$ aws s3 sync public s3://www.gofaas.net/
$ aws cloudfront create-invalidation --distribution-id E2YL0GMGANCGMA --paths '/*'
```

Sure enough we can access our content via the CDN:

```shell
$ curl https://d2bwnae7bzw1t6.cloudfront.net/
...
<title>My first gofaas/Vue app</title>
```

## DNS

The final step is to set up a DNS CNAME from our `WebDomainName` parameter (e.g. `www.gofaas.net`) to the new `WebDistributionDomainName` output (e.g. `d2bwnae7bzw1t6.cloudfront.net`).

If we are using Route53, this is easy to do through the UI:

<p align="center"><img src="img/route53.png" alt="alt text" width="410" /></p>

In this case we could consider automating DNS setup by adding an conditional `AWS::Route53::RecordSet` resource to our template. We could also consider using [ACM DNS validation](https://docs.aws.amazon.com/acm/latest/userguide/gs-acm-validate-dns.html) to fully automate the certificate.

After a few minutes we have our custom HTTPS endpoint:

```shell
$ curl https://www.gofaas.net
...
<p>Hello world! This is HTML5 Boilerplate.</p>
```

## Summary

When hosting a static site an app with S3, CloudFront and ACM we can:

- Store our static web content for low cost
- Access cached web content quickly via a custom domain
- Automate cert creation and renewal

We no longer have to worry about:

- Configuring HTTP servers
- Generating HTML content in our API

Our app is easier to build and more reliable and cost effective to run.