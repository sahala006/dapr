from django.http import HttpResponse
from django.shortcuts import render
import time
import socket

def home(request):
    print (request.body)
    hostname = socket.gethostname()    
    return HttpResponse(time.strftime(hostname + ' ' + '%Y-%m-%d %H:%M:%S',time.localtime(time.time())) + '\n')

