# lab-python-runner
# Python runner image for lab exercise evaluation.
# Based on python:3.11-slim.
# Adds make and bash so Makefile entrypoints work correctly, along with data science packages (pandas, matplotlib).

FROM python:3.11-slim
RUN apt-get update && apt-get install -y --no-install-recommends make bash && rm -rf /var/lib/apt/lists/*
RUN pip install --no-cache-dir pandas matplotlib
