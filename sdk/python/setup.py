from setuptools import setup, find_packages

setup(
    name="valt",
    version="0.1.0",
    packages=find_packages(),
    python_requires=">=3.8",
    description="Valt secret vault Python SDK",
    long_description="A minimal Python client for the Valt secret vault API.",
    author="Valt",
    license="MIT",
    package_data={"valt": ["py.typed"]},
)
