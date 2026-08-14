# DESIGN TRADEOFFS

## Architecture

The application uses a DDD-lite inspired architecture to make it well organized, structured and modular.

Each domain worries about and implements it's own handlers, services, repositories, models, and other component like transaction runner and unit of work where applicable.

Repositories are abstracted using interfaces, with implementations created for specific database types (eg pgsql, mysql, sqlite). Services depend on these interfaces, this way, repo implementation can be swapped out with minimal code change. This also improve testability as the repositories can be mocked to test services that depend on them.

The platform package holds implementations that are not domain related, they fall under two categories:

1. Abstractions that the application depends on to start, like the `database` package which bootstraps postgresql and redis, and the `config` package which abstracts environment variables and configurations needed to run the app.
2. data transfer layers like `notifier` which let's packages send and receive data in realtime and the `render` package which abstracts the handling of rendering of json data and pages.

## State management

There are two different states that are being managed. Each answering different questins:

1. What is the current state or status of the payment?

This first question is answered by having a `status` columns on the `payment_intents` table that tracks the current state which the payment is in. The initial state of all PaymentIntent is `pending` with the final state being one of `captured`, `refunded`, `cancled`, or `failed`.  
The status does not update until it has been comfirmed by the bank that the action has succeeded. For example, a payment does not go from captured to refunded until it has been refund has been confirmed by the bank.

2. What series of states did the payment go trough to get to it's current state.

The second question is answered by having another table, `payment_events`, which has a many-to-one relationship with `payment_intent`.  
Because `payment_intents.status` only tracks the current state, which causes the history of the payment to be lost when it moves to the next state, a PaymentEvent entry is created when the transition happens. So when a paymenet moves from `pending` to `authorized`, a PaymentEvent entry is created with `state=authorized` and `external_state_id` set to the state _id_ returned by the bank (authorization_id, catured_id, etc). So a payment or PaymentIntent that was refunded will have it's `status` field set as `refunded` and three PaymentEvent entries, each one with `state` set to authorized, captured and refunded. And using the `created_at` field, this PaymentEvent entries can be used by the UI client to display the payment history or reciept as needed.

## Failure handling.

The retry strategy is coupled with the message queue system for simplicity.

The errors from the bank is categorized into

1. Terminal error: errors that always stay true no matter how many times it's retried
2. Transient error - errors that are temporary and have the possibility of being resolved on subsequent retries. Examples include errors caused by network issues, service outage or system downtime.

Only transient errors are retried, retries happens threee times (configurable), with exponential backoff. The process is recorded as failed if the final retry fails.

The system works in a way that the client gets a response within 30 seconds of request. If a transient error is flagged, the system keeps retrying but returns a response to the client within a designated time (in this case 30secs).

The `payment_intents` table has a `current_process` column which is set to a process (authorizing, capturing, refunding or cancelling) before it starts and reset back to null when the process ends. So if a process is still running (maybe due to retries) the client gets the `current_operation` field along with the reponse which allows it to know that an operation is runnnig and which operation it is.  
For example, when a payment capture request is made for an already authorized PaymentIntent, `payment_intents.current_operation` is set to `capturing`. if the bank does not captures the payment within the timeout window (30secs) the `current_operation` field stays unchanged and the `status` field does not get updated so the client gets:

```
{
    "status": "authorized",
    "current_operation": "capturing",
}
```

On the other hand, if the payment is captured before the timeout window, the `current_operation` field is reset to `null` and the `status` is updated to `authorized` and so the client gets:

```
{
    "status": "captured",
    "current_operation": "",
}
```

> _**N.B:**_ After an operation is completed, the client can get the updated data via websocket, webhook or client polling. That is beyond the scope of this project at the moment and is left unimplemented. The main point here is to not leave the client hanging while the system is retrying a transient error.

## Idempotency

The system is completely idempotent. non-get requests requires a `X-Idempotency-Key` header which is client generated.

The system stores the idempotency key on redis as a key-value data with the idempotency key being the key and a stringified object of the request hash and the payment reference being the value. This way if a different request is made with an idempotency key that already exists on the system, it is rejected.  
One tradeoff to this design though it that stored idempotency key entries needs to be cleared occasionally to prevent the redis store from getting bloated with irrelevant idempotency key.

When the same request is sent with the same idempotency key, the system replays the reponse and adds a response header `X-Idempotent-Replayed=true`

## What you'd do differently (With more time or in prosduction, what would change?)

Right now, if for whatever reason the messaging system goes down, there is no recovery, and so not only will the processes not run, they also will be lost, leaving a pending (unauthorized) payment to remain pending with no way to reactivate the process of capturing it.

The first thing I would do if I were to continue work on this project is to introduce 'outbox pattern'. That way the processes are written in the outbox before it is promoted to the queue system, if the queus system fails, stale processes can be takebn from the outbox and continued.

Another thing is a better logging system. Right now at every junction where there can be errors, the language loging system is used to log the error detail on the teninal, a more useful way might be to write the errors associated with a payment intent inside a log file, that way it can be parsed latter and dislayed to support or stakeholders.
