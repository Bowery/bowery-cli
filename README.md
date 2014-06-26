# Introducing Bowery
_Welcome to the Future of Web Development._

## Philosophy
Web development is terrible—it's filled with convoluted abstrations that have
little to do with building a product. We spend most of our time worrying about
server and database management, when we should be spending it building the product.

When building your product with Bowery, you write the application and the rest
is taken care of for you—there's no development environment, no databases you
have to manage, etc. Common services that are needed are easily added, using
Bowery. Implementation details and performance optimizations are made behind
the scenes so that you can focus all of your energy on the product.

## CLI

### Installation

# TODO, FILL OUT INSTALLATION DETAILS

### Signup

Before you can start building your application, you'll need to signup for Bowery.

```
bowery signup
```

You'll be given a few prompts to input some information needed for your account,
and afterwards you'll be logged in and ready to go.

### Managing services

Now that you've signed up you can prepare the services you need to run your
application.

You can add services by running the following

```
bowery add <your service names>
```

This will add the services and ask you for the type of the services and a path
if needed.

You can of course remove services by using

```
bowery add <your service names>
```

This will remove the listed service names.

If you want to know the services you have currently, use the following command

```
bowery info
```

You don't have to use these commands though, when adding a service it just adds
them to a `bowery.json` file in the director, so if you want you can edit it
directly instead.

### Connecting

So you've created the services you want for your application, and wrote some code
and are ready to deploy it and check it out. You can connect and sync changes
using the following command

```
bowery connect
```

When you connect all changes you make to a services code are uploaded to the
service and you can see changes immediately.

### Managing your application

After you've connected, there's some extra things included to make it easy to
manage and keep track of the services.

If you need to access the service for maintenance or whatever reason you can
connect to it via SSH.

```
bowery ssh <service name>
```

When you're connected you also receive logs for the services as they come in,
use the following command to read them

```
bowery logs
```
