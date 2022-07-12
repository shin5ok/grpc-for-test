import sys
sys.path.append("../../")
sys.path.append("../../pb")
import grpc
import os
import pb.simple_pb2 as pb2
import pb.simple_pb2_grpc as pb2_grpc
from pydantic import BaseModel
import typing
import click

grpc_host = os.environ.get("GRPC_HOST")
insecure = os.environ.get("INSECURE")

if insecure:
    channel = grpc.insecure_channel(grpc_host)
else:
    channel = grpc.secure_channel(grpc_host, grpc.ssl_channel_credentials())
stub = pb2_grpc.SimpleStub(channel)

@click.group()
@click.option("--insecure", is_flag=False)
def cli(insecure: bool):
    ...

@cli.command()
def put_message() -> None:
    name = pb2.Name(id=1, text="tako")
    message = pb2.Message(name=name, message="takosuke")

    results = stub.PutMessage(message)

    print(results)

@cli.command()
@click.option("--number", type=int, default=1)
def list_message(number: int) -> None:
    request = pb2.Request(number=number)
    r = stub.ListMessage(request)
    for x in r:
        print(x, end='')

if __name__ == '__main__':
    cli()
