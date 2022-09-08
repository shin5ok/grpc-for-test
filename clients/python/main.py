import sys
sys.path.append("../../")
sys.path.append("../../pb")
import grpc
import os
import pb.simple_pb2 as pb2
import pb.simple_pb2_grpc as pb2_grpc
import typing
import click
import datetime

grpc_host: str = os.environ.get("GRPC_HOST")
insecure: str = os.environ.get("INSECURE")

if insecure:
    channel = grpc.insecure_channel(grpc_host)
else:
    channel = grpc.secure_channel(grpc_host, grpc.ssl_channel_credentials())
stub = pb2_grpc.SimpleStub(channel)

@click.group()
def cli():
    ...

@cli.command()
@click.option("--number", type=int, default=1)
@click.option("--stdout", is_flag=True, default=False)
@click.argument("value", default="foo")
def put_message(number: int, stdout: bool, value: str) -> None:
    start = datetime.datetime.now()
    for id in range(1, number):
        name = pb2.Name(id=id, text=value)
        message = pb2.Message(name=name, message="tako")

        results = stub.PutMessage(message)
        if stdout:
            print(results, end='')

    finish = datetime.datetime.now()
    delta = finish - start
    print(f"{delta.total_seconds()}\n")


@cli.command()
@click.option("--number", type=int, default=1)
@click.option("--stdout", is_flag=True, default=False)
def list_message(number: int, stdout: bool) -> None:
    start = datetime.datetime.now()
    request = pb2.Request(number=number)
    r = stub.ListMessage(request)
    for x in r:
        value = x.message
        if stdout:
            print(value, end='\n')
    finish = datetime.datetime.now()
    delta = finish - start
    print(f"{delta.total_seconds()}\n")

if __name__ == '__main__':
    cli()
